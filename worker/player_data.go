package worker

import (
	"context"
	"dota_league/api"
	"dota_league/model"
	"log/slog"
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

	hasRecord, err := dl.PlayerRepository.HasAnyRecord(ctx)
	if err != nil {
		return nil, err
	}

	if hasRecord {
		for _, player := range playersData.Players {
			//TODO: we need to update values sometimes

			//Store player only if it not exists in the DB
			b, err := dl.PlayerRepository.ExistsByID(ctx, player.ID)
			if err != nil {
				return nil, err
			}

			if !b {
				if err = dl.PlayerRepository.Store(ctx, &player); err != nil {
					return nil, err
				}
			}
		}

	} else {
		err = dl.PlayerRepository.StoreAll(ctx, playersData.Players)
		if err != nil {
			return nil, err
		}
	}

	return playersData.Players, nil
}

func (dl *DataLoader) storeSinglePlayer(ctx context.Context, playerID int) error {
	player, err := api.LoadSinglePlayer(ctx, playerID)
	if err != nil {
		return err
	}

	err = dl.PlayerRepository.Store(ctx, player)
	if err != nil {
		return err
	}

	return nil
}
