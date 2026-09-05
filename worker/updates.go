package worker

import (
	"context"
	"dota_league/model"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strconv"
)

type UpdatesStore interface {
	Store(context.Context, *model.Update) error
}

// recordUpdate receives only explicitly selected public fields, never raw API
// payloads or server errors. Refresh timestamps and DB keys are excluded.
func (dl *DataLoader) recordUpdate(ctx context.Context, entity string, id int, name string, before, after map[string]any) {
	changes := []model.UpdateChange{}
	action := "created"
	if before != nil {
		action = "updated"
		for _, field := range slices.Sorted(maps.Keys(after)) {
			if !reflect.DeepEqual(before[field], after[field]) {
				changes = append(changes, model.UpdateChange{Field: field, Before: before[field], After: after[field]})
			}
		}
		if len(changes) == 0 {
			return
		}
	}
	url := ""
	switch entity {
	case "tournament":
		url = "/league/" + strconv.Itoa(id)
	case "team", "roster":
		url = "/team/" + strconv.Itoa(id)
	}
	update := &model.Update{Entity: entity, EntityID: id, Name: name, Action: action, URL: url, Changes: changes}
	if entity == "roster" {
		update.RosterAdmins, _ = after["admins"].(map[int]bool)
	}
	dl.storeUpdate(ctx, update)
}

func (dl *DataLoader) recordPlayerUpdate(ctx context.Context, player *model.Player) {
	team := &model.UpdateTeam{ID: player.TeamID, Name: player.TeamName}
	if team.ID > 0 && team.Name == "" && dl.TeamRepository != nil {
		if stored, err := dl.TeamRepository.GetByID(ctx, team.ID); err == nil && stored != nil {
			team.Name = stored.Name
		}
	}
	dl.storeUpdate(ctx, &model.Update{Entity: "player", EntityID: player.ID, Name: player.Name, Action: "created", Team: team, Changes: []model.UpdateChange{}})
}

func (dl *DataLoader) storeUpdate(ctx context.Context, update *model.Update) {
	if err := dl.Updates.Store(ctx, update); err != nil {
		// Feed availability must not turn a completed source-data write into a failure.
		slog.ErrorContext(ctx, "record update", "entity", update.Entity, "entity_id", update.EntityID, "error", err)
	}
}

func teamSnapshot(team *model.Team) map[string]any {
	if team == nil {
		return nil
	}
	return map[string]any{
		"name": team.Name, "tag": team.Tag, "region": team.Region,
		"country_code": team.CountryCode, "url": team.URL, "url_logo": team.URLLogo,
		"wins": team.Wins, "losses": team.Losses, "team_captain": team.TeamCaptain,
	}
}

func tournamentSnapshot(league *model.LeagueDetails) map[string]any {
	if league == nil {
		return nil
	}
	return map[string]any{
		"name": league.Name, "tier": league.Tier, "region": league.Region,
		"start_timestamp": league.StartTimestamp, "end_timestamp": league.EndTimestamp,
		"status": league.Status, "total_prize_pool": league.TotalPrizePool,
		"description": league.Description, "url": league.URL,
	}
}

func rosterSnapshot(roster *model.TeamRoster) map[string]any {
	if roster == nil {
		return nil
	}
	members := make(map[int]bool, len(roster.TeamMembers))
	admins := make(map[int]bool)
	for _, member := range roster.TeamMembers {
		members[member.AccountID] = member.IsActive
		if member.Admin {
			admins[member.AccountID] = true
		}
	}
	return map[string]any{"members": members, "admins": admins}
}
