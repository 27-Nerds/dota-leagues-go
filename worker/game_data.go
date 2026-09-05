package worker

import (
	"context"
	"dota_league/api"
)

func (dl *DataLoader) performGamesUpdate(ctx context.Context) error {
	liveGames, err := api.LoadLiveGames(ctx)
	if err != nil {
		return err
	}
	previousGames, err := dl.GameRepository.GetAll(ctx)
	if err != nil {
		return err
	}

	activeLeagues := make(map[int]bool)
	for _, game := range liveGames.Games {
		activeLeagues[game.LeagueID] = true
		dl.LiveGamesManager.AddGame(game)
	}
	finishedLeagues := make(map[int]bool)
	for _, game := range previousGames {
		if !activeLeagues[game.LeagueID] {
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
		if err := dl.LeagueDetailsRepository.UpdateLiveStatus(ctx, leagueID, false); err != nil {
			return err
		}
	}

	// An empty upstream snapshot must also clear games that have finished.
	if err := dl.GameRepository.RemoveAll(ctx); err != nil {
		return err
	}
	if len(liveGames.Games) == 0 {
		return nil
	}
	return dl.GameRepository.StoreAll(ctx, liveGames.Games)
}
