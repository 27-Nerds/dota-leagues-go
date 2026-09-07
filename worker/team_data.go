package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"dota_league/model"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

// storeTeam gets data from api and stores or updates it in the DB.
// Teams refreshed within detailsRefreshInterval are skipped to spare the Valve API.
func (dl *DataLoader) storeTeam(ctx context.Context, teamID int) error {

	exist, err := dl.TeamRepository.ExistsByID(ctx, teamID)
	if err != nil {
		return err
	}

	var stored *model.Team
	if exist {
		var gerr error
		stored, gerr = dl.TeamRepository.GetByID(ctx, teamID)
		switch {
		case gerr != nil:
			return gerr
		case time.Since(time.Unix(stored.UpdatedTimestamp, 0)) < detailsRefreshInterval:
			err := dl.downloadTeamImage(ctx, stored)
			if e.IsNotFound(err) {
				return nil
			}
			return err
		}
	}

	team, err := api.LoadTeamDetails(ctx, teamID)
	if err != nil {
		return err
	}

	if err := dl.storeTeamRoster(ctx, team, exist); err != nil {
		return err
	}

	team.UpdatedTimestamp = time.Now().Unix()

	if exist {
		err = dl.TeamRepository.Update(ctx, team)
	} else {
		err = dl.TeamRepository.Store(ctx, team)
	}
	if err != nil {
		return err
	}

	dl.recordUpdate(ctx, "team", team.ID, team.Name, teamSnapshot(stored), teamSnapshot(team))

	err = dl.downloadTeamImage(ctx, team)
	if err != nil && !e.IsNotFound(err) {
		slog.WarnContext(ctx, "download team image", "team_id", teamID, "error", err)
	}

	return nil
}

func (dl *DataLoader) downloadTeamImage(ctx context.Context, team *model.Team) error {

	path := filepath.Join(assetDirectory(), "teams", fmt.Sprint(team.ID))
	if team.URLLogo != "" {
		err := api.DownloadImageIfNotExist(ctx, team.URLLogo, path, "logo.png")
		if !e.IsNotFound(err) {
			return err
		}
	}

	// Valve's website also serves team logos by ID when the API URL is absent.
	url := fmt.Sprintf("https://cdn.steamstatic.com/apps/dota2/teamlogos/%d.png", team.ID)
	return api.DownloadImageIfNotExist(ctx, url, path, "logo.png")
}

func (dl *DataLoader) storeTeamRoster(ctx context.Context, team *model.Team, teamExists bool) error {
	teamRoster := model.TeamRoster{
		TeamID: team.ID,
	}

	for _, member := range team.Members {
		// valve returns placeholder members (account_id 0) for empty roster slots
		if member.AccountID == 0 {
			continue
		}

		needsRefresh, err := dl.PlayerRepository.NeedsProfileRefresh(ctx, member.AccountID, time.Now())
		if err != nil {
			return err
		}
		if needsRefresh {
			slog.DebugContext(ctx, "queue team member", "account_id", member.AccountID)
			if err := enqueue(ctx, dl.LoadSinglePlayer, member.AccountID); err != nil {
				return err
			}
		}

		teamMember := model.TeamMember{
			AccountID: member.AccountID,
			IsActive:  true,
			Admin:     member.Admin,
		}
		teamRoster.TeamMembers = append(teamRoster.TeamMembers, teamMember)
	}

	exist, err := dl.TeamRosterRepository.ExistsByTeamID(ctx, team.ID)
	if err != nil {
		return err
	}

	var stored *model.TeamRoster
	if exist {
		stored, err = dl.TeamRosterRepository.GetByID(ctx, team.ID)
		if err != nil {
			return err
		}
		err = dl.TeamRosterRepository.Update(ctx, &teamRoster)
	} else {
		err = dl.TeamRosterRepository.Store(ctx, &teamRoster)
	}
	if err != nil {
		return err
	}

	// Initial membership belongs to the team creation event. Publish roster
	// activity only when an existing team's recorded membership changes.
	if teamExists && stored != nil {
		dl.recordUpdate(ctx, "roster", team.ID, team.Name, rosterSnapshot(stored), rosterSnapshot(&teamRoster))
		dl.recordRosterMoves(ctx, team, stored, &teamRoster)
	}
	return nil
}

// recordRosterMoves mirrors a roster change onto each player who joined or left, so a
// player's own log shows team history even when the DPC feed reports no team.
func (dl *DataLoader) recordRosterMoves(ctx context.Context, team *model.Team, before, after *model.TeamRoster) {
	was := make(map[int]bool, len(before.TeamMembers))
	for _, m := range before.TeamMembers {
		was[m.AccountID] = true
	}
	now := make(map[int]bool, len(after.TeamMembers))
	for _, m := range after.TeamMembers {
		now[m.AccountID] = true
	}
	var moved []int
	for id := range now {
		if !was[id] {
			moved = append(moved, id)
		}
	}
	for id := range was {
		if !now[id] {
			moved = append(moved, id)
		}
	}
	if len(moved) == 0 {
		return
	}
	slices.Sort(moved)
	names, err := dl.PlayerRepository.GetProfiles(ctx, moved)
	if err != nil {
		slog.WarnContext(ctx, "resolve roster move names", "team_id", team.ID, "error", err)
	}
	teamName := team.Name
	if teamName == "" {
		teamName = "Team #" + strconv.Itoa(team.ID)
	}
	for _, id := range moved {
		name := names[id].Name
		if name == "" {
			name = names[id].SteamName
		}
		if name == "" {
			name = "Player #" + strconv.Itoa(id)
		}
		before := map[string]any{"roster_team": "", "roster_team_id": 0}
		after := map[string]any{"roster_team": teamName, "roster_team_id": team.ID}
		if !now[id] {
			before, after = after, before
		}
		dl.recordUpdate(ctx, "player", id, name, before, after)
	}
}

// Scan all stored teams independently of the current player directory. The
// single consumer applies the freshness guard again when each job reaches it.
func (dl *DataLoader) performTeamsUpdate(ctx context.Context) error {
	ids, err := dl.TeamRepository.GetRefreshCandidates(ctx, time.Now())
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := enqueue(ctx, dl.LoadTeam, id); err != nil {
			return err
		}
	}
	return nil
}
