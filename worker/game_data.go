package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"dota_league/model"
)

func (dl *DataLoader) performGamesUpdate(ctx context.Context) error {
	liveGames, err := api.LoadLiveGames(ctx)
	if err != nil {
		return err
	}
	return dl.applyGamesUpdate(ctx, liveGames.Games)
}

func (dl *DataLoader) applyGamesUpdate(ctx context.Context, games []model.Game) error {
	previousGames, err := dl.GameRepository.GetAll(ctx)
	if err != nil {
		return err
	}

	activeLeagues := make(map[int]bool)
	for _, game := range games {
		if game.LeagueID > 0 {
			activeLeagues[game.LeagueID] = true
		}
	}
	dl.LiveGamesManager.ReplaceGames(games)
	finishedLeagues := make(map[int]bool)
	for _, game := range previousGames {
		if game.LeagueID > 0 && !activeLeagues[game.LeagueID] {
			finishedLeagues[game.LeagueID] = true
		}
	}

	for leagueID := range activeLeagues {
		exists, err := dl.LeagueDetailsRepository.ExistsByID(ctx, leagueID)
		if err != nil {
			return err
		}
		if exists {
			if err := dl.LeagueDetailsRepository.UpdateLiveStatus(ctx, leagueID, true); err != nil {
				return err
			}
		} else if err := enqueue(ctx, dl.LoadLeagueDetails, leagueID); err != nil {
			return err
		}
	}
	for leagueID := range finishedLeagues {
		if err := dl.LeagueDetailsRepository.UpdateLiveStatus(ctx, leagueID, false); err != nil && !e.IsNotFound(err) {
			return err
		}
	}

	// An empty upstream snapshot must also clear games that have finished.
	if err := dl.GameRepository.RemoveAll(ctx); err != nil {
		return err
	}
	if len(games) == 0 {
		return nil
	}
	return dl.GameRepository.StoreAll(ctx, games)
}
