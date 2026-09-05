package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"fmt"
	"log/slog"
	"time"
)

func (dl *DataLoader) downloadLeagueImage(ctx context.Context, leagueID int) error {
	// Use the shared CDN path used by the Dota client. The older website
	// path under apps/dota2/images/leagues does not contain newer leagues.
	url := fmt.Sprintf("https://shared.steamstatic.com/dota_leagues/images/%d/image_8.png", leagueID)

	path := fmt.Sprintf("public/%d", leagueID)

	err := api.DownloadImageIfNotExist(ctx, url, path, "logo.png")
	if e.IsNotFound(err) {
		// if file image_8 not found on the dota 2 server, try to redownload image_1
		url = fmt.Sprintf("https://shared.steamstatic.com/dota_leagues/images/%d/image_1.png", leagueID)

		err := api.DownloadImageIfNotExist(ctx, url, path, "logo.png")
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	return nil
}

func (dl *DataLoader) performPrizePoolUpdate(ctx context.Context) error {
	slog.DebugContext(ctx, "refreshing prize pools")

	// update prizepool only for active tier 4 and 5 leagues
	leagues, err := dl.LeagueDetailsRepository.GetAllActiveForTiers(ctx, []int{4, 5})
	if err != nil {
		return err
	}

	for _, league := range leagues {
		prizePool, err := api.LoadPrizePool(ctx, league.ID)
		if err != nil {
			return err
		}
		// do not store 0 prizepool
		if prizePool.PrizePool == 0 {
			continue
		}

		err = dl.LeagueDetailsRepository.UpdateTotalPrizePool(ctx, league.ID, prizePool.PrizePool)
		if err != nil {
			return err
		}
	}

	return err
}

func (dl *DataLoader) performLeaguesUpdate(ctx context.Context) error {

	// Update leagues in the DB
	err := dl.storeLeagues(ctx)
	if err != nil {
		return err
	}
	slog.DebugContext(ctx, "base league information updated")

	//TODO: store last processed league id to not
	//      Get all fresh leagues
	leagues, err := dl.LeagueRepository.GetAllActive(ctx)
	if err != nil {
		return err
	} else {
		slog.DebugContext(ctx, "refreshing league details")
		for _, league := range leagues {
			if err := enqueue(ctx, dl.LoadLeagueDetails, league.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// storeLeagues gets data from api and stores it into DB
func (dl *DataLoader) storeLeagues(ctx context.Context) error {
	leagueData, err := api.LoadLeagues(ctx)
	if err != nil {
		return err
	}

	hasRecord, err := dl.LeagueRepository.HasAnyRecord(ctx)
	if err != nil {
		return err
	}

	if hasRecord {
		for _, league := range leagueData.Leagues {
			//Store league only if it not exists in the DB
			b, err := dl.LeagueRepository.ExistsByID(ctx, league.ID)
			if err != nil {
				return err
			}

			if !b {
				if err = dl.LeagueRepository.Store(ctx, &league); err != nil {
					return err
				}
			}
		}

	} else {
		err = dl.LeagueRepository.StoreAll(ctx, leagueData.Leagues)
		if err != nil {
			return err
		}
	}

	return nil
}

// storeLeagueDetails gets data from api and stores it into DB (upsert + series refresh).
// Records refreshed within detailsRefreshInterval are skipped to spare the Valve API.
func (dl *DataLoader) storeLeagueDetails(ctx context.Context, leagueID int) error {
	exist, err := dl.LeagueDetailsRepository.ExistsByID(ctx, leagueID)
	if err != nil {
		return err
	}

	if exist {
		stored, gerr := dl.LeagueDetailsRepository.GetByID(ctx, leagueID)
		switch {
		case gerr != nil:
			return gerr
		case time.Since(time.Unix(stored.UpdatedTimestamp, 0)) < detailsRefreshInterval:
			// A successful metadata refresh does not imply the logo downloaded.
			return dl.downloadLeagueImage(ctx, leagueID)
		}
	}

	leagueDetails, err := api.LoadLeagueDetails(ctx, leagueID)
	if err != nil {
		return err
	}

	// add pricepool info to details struct
	leagueDetails.Details.BasePrizePool = leagueDetails.PrizePool.BasePrizePool
	leagueDetails.Details.TotalPrizePool = leagueDetails.PrizePool.TotalPrizePool

	// add stream info
	leagueDetails.Details.Streams = leagueDetails.Streams

	if err := dl.LeagueSeriesRepository.ReplaceAllForLeague(ctx, leagueID, leagueDetails.SeriesInfos); err != nil {
		return err
	}

	// queue teams referenced by the series but missing from the DB so schedule names can be joined later
	seen := make(map[int]bool)
	for _, s := range leagueDetails.SeriesInfos {
		for _, teamID := range []int{s.TeamID1, s.TeamID2} {
			if teamID == 0 || seen[teamID] {
				continue
			}
			seen[teamID] = true

			exists, err := dl.TeamRepository.ExistsByID(ctx, teamID)
			if err != nil {
				// transient DB failure is not a missing team - do not fan out an api request for it
				slog.WarnContext(ctx, "check series team", "team_id", teamID, "error", err)
				continue
			}

			if !exists {
				if err := enqueue(ctx, dl.LoadTeam, teamID); err != nil {
					return err
				}
			}
		}
	}

	leagueDetails.Details.UpdatedTimestamp = time.Now().Unix()
	if exist {
		err = dl.LeagueDetailsRepository.Update(ctx, &leagueDetails.Details)
	} else {
		err = dl.LeagueDetailsRepository.Store(ctx, &leagueDetails.Details)
	}
	if err != nil {
		return err
	}

	err = dl.downloadLeagueImage(ctx, leagueID)
	if err != nil {
		slog.WarnContext(ctx, "download league image", "league_id", leagueID, "error", err)
	}

	return nil
}
