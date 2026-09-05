package worker

import (
	"context"
	"dota_league/api"
	"dota_league/model"
	"fmt"
	"log/slog"
	"time"
)

// storeTeam gets data from api and stores or updates it in the DB.
// Teams refreshed within detailsRefreshInterval are skipped to spare the Valve API.
func (dl *DataLoader) storeTeam(ctx context.Context, teamID int) error {

	exist, err := dl.TeamRepository.ExistsByID(ctx, teamID)
	if err != nil {
		return err
	}

	if exist {
		stored, gerr := dl.TeamRepository.GetByID(ctx, teamID)
		switch {
		case gerr != nil:
			return gerr
		case time.Since(time.Unix(stored.UpdatedTimestamp, 0)) < detailsRefreshInterval:
			return nil
		}
	}

	team, err := api.LoadTeamDetails(ctx, teamID)
	if err != nil {
		return err
	}

	if err := dl.storeTeamRoster(ctx, team); err != nil {
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

	err = dl.downloadTeamImage(ctx, team)
	if err != nil {
		slog.WarnContext(ctx, "download team image", "team_id", teamID, "error", err)
	}

	return nil
}

func (dl *DataLoader) downloadTeamImage(ctx context.Context, team *model.Team) error {

	// skip if logo url is empty
	if team.URLLogo == "" {
		return nil
	}

	path := fmt.Sprintf("public/teams/%d", team.ID)

	err := api.DownloadImageIfNotExist(ctx, team.URLLogo, path, "logo.png")
	if err != nil {
		return err
	}

	return nil
}

func (dl *DataLoader) storeTeamRoster(ctx context.Context, team *model.Team) error {
	teamRoster := model.TeamRoster{
		TeamID: team.ID,
	}

	for _, member := range team.Members {
		// valve returns placeholder members (account_id 0) for empty roster slots
		if member.AccountID == 0 {
			continue
		}

		exist, err := dl.PlayerRepository.ExistsByID(ctx, member.AccountID)
		if err != nil {
			return err
		}
		if !exist {
			slog.DebugContext(ctx, "queue team member", "account_id", member.AccountID)
			if err := enqueue(ctx, dl.LoadSinglePlayer, member.AccountID); err != nil {
				return err
			}
		}

		teamMember := model.TeamMember{
			AccountID: member.AccountID,
			IsActive:  true,
		}
		teamRoster.TeamMembers = append(teamRoster.TeamMembers, teamMember)
	}

	exist, err := dl.TeamRosterRepository.ExistsByTeamID(ctx, team.ID)
	if err != nil {
		return err
	}

	if exist {
		err = dl.TeamRosterRepository.Update(ctx, &teamRoster)
	} else {
		err = dl.TeamRosterRepository.Store(ctx, &teamRoster)
	}
	if err != nil {
		return err
	}

	return nil
}
