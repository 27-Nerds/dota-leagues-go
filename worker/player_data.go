package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"time"
)

func (dl *DataLoader) performPlayersUpdate(ctx context.Context) error {
	players, err := dl.storePlayers(ctx)
	if err != nil {
		return err
	}
	slog.DebugContext(ctx, "player information updated")
	seen := make(map[int]bool)
	for _, player := range players {
		// skip players with team id = 0
		if player.TeamID == 0 || seen[player.TeamID] {
			continue
		}

		seen[player.TeamID] = true
		// Existing teams have their own refresh sweep, including empty rosters.
		exists, err := dl.TeamRepository.ExistsByID(ctx, player.TeamID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := enqueue(ctx, dl.LoadTeam, player.TeamID); err != nil {
			return err
		}
	}

	return nil
}

// storePlayers gets data from api and stores it into DB
func (dl *DataLoader) storePlayers(ctx context.Context) ([]model.Player, error) {
	playersData, err := api.LoadPlayers(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(playersData.Players))
	for _, player := range playersData.Players {
		ids = append(ids, player.ID)
	}
	stored, err := dl.PlayerRepository.GetProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, player := range playersData.Players {
		created, err := dl.PlayerRepository.SaveProfile(ctx, &player)
		if err != nil {
			return nil, err
		}
		if created {
			dl.recordPlayerUpdate(ctx, &player)
			continue
		}
		if previous, ok := stored[player.ID]; ok {
			// A Steam fallback gaining a DPC profile is an identity change worth showing,
			// so the fallback's name is compared like any other stored value.
			dl.recordUpdate(ctx, "player", player.ID, player.Name, playerSnapshot(&previous), playerSnapshot(&player))
		}
	}

	return playersData.Players, nil
}

func (dl *DataLoader) storeSinglePlayer(ctx context.Context, playerID int) error {
	return dl.storeSinglePlayerWithLoader(ctx, playerID, api.LoadPlayerWithSteamFallback)
}

const missingPlayerRetryInterval = 24 * time.Hour
const missingPlayerCacheLimit = 4096

func (dl *DataLoader) storeSinglePlayerWithLoader(ctx context.Context, playerID int, load func(context.Context, int) (*model.Player, error)) error {
	needsRefresh, err := dl.PlayerRepository.NeedsProfileRefresh(ctx, playerID, time.Now())
	if err != nil {
		return err
	}
	if !needsRefresh {
		delete(dl.missingPlayers, playerID)
		return nil
	}
	if time.Now().Before(dl.missingPlayers[playerID]) {
		return nil
	}
	delete(dl.missingPlayers, playerID)
	player, err := load(ctx, playerID)
	if e.IsNotFound(err) && ctx.Err() == nil {
		dl.rememberMissingPlayer(playerID)
		slog.DebugContext(ctx, "player profile unavailable from DPC and Steam", "account_id", playerID, "retry_after", dl.missingPlayers[playerID])
		return nil
	}
	if err != nil {
		return err
	}

	created, err := dl.PlayerRepository.SaveProfile(ctx, player)
	if err != nil {
		return err
	}
	if created {
		dl.recordPlayerUpdate(ctx, player)
	}

	return nil
}

// Keep negative lookups bounded; expired entries are removed on the next miss.
func (dl *DataLoader) rememberMissingPlayer(playerID int) {
	now := time.Now()
	if dl.missingPlayers == nil {
		dl.missingPlayers = make(map[int]time.Time)
	}
	oldestID := 0
	var oldest time.Time
	for id, until := range dl.missingPlayers {
		if !now.Before(until) {
			delete(dl.missingPlayers, id)
			continue
		}
		if oldest.IsZero() || until.Before(oldest) {
			oldestID, oldest = id, until
		}
	}
	if len(dl.missingPlayers) >= missingPlayerCacheLimit {
		delete(dl.missingPlayers, oldestID)
	}
	dl.missingPlayers[playerID] = now.Add(missingPlayerRetryInterval)
}
