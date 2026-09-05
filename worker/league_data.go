package worker

import (
	"dota_league/api"
	e "dota_league/error"
	"fmt"
	"log"
	"time"
)

func (dl *DataLoader) downloadLeagueImage(leagueID int) error {
	url := fmt.Sprintf("http://cdn.dota2.com/apps/dota2/images/leagues/%d/images/image_8.png", leagueID)

	path := fmt.Sprintf("public/%d", leagueID)

	err := api.DownloadImageIfNotExist(url, path, "logo.png")
	if e.IsNotFound(err) {
		// if file image_8 not found on the dota 2 server, try to redownload image_1
		url = fmt.Sprintf("http://cdn.dota2.com/apps/dota2/images/leagues/%d/images/image_1.png", leagueID)

		err := api.DownloadImageIfNotExist(url, path, "logo.png")
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	return nil
}

func (dl *DataLoader) performPrizePoolUpdate() error {
	log.Println("refreshing prizepools...")

	// update prizepool only for active tier 4 and 5 leagues
	leagues, err := dl.LeagueDetailsRepository.GetAllActiveForTiers([]int{4, 5})
	if err != nil {
		return err
	}

	for _, league := range *leagues {
		prizePool, err := api.LoadPrizePool(league.ID)
		if err != nil {
			log.Printf("error while performPrizePoolUpdate - LoadPrizePool: %v", err)
			return err
		}
		// do not store 0 prizepool
		if prizePool.PrizePool == 0 {
			log.Printf("0 prizepool recieved for league %d", league.ID)
			return nil
		}

		err = dl.LeagueDetailsRepository.UpdateTotalPrizePool(league.ID, prizePool.PrizePool)
		if err != nil {
			log.Printf("error while performPrizePoolUpdate - UpdateTotalPrizePool: %v", err)
			return err
		}
	}

	return err
}

func (dl *DataLoader) performLeaguesUpdate() error {

	// Update leagues in the DB
	err := dl.storeLeagues()
	if err != nil {
		log.Printf("error while storeLeagues: %v", err)
		return err
	}
	log.Println("base info for leagues updated.")

	//TODO: store last processed league id to not
	//      Get all fresh leagues
	leagues, err := dl.LeagueRepository.GetAllActive()
	if err != nil {
		log.Printf("error while storeLeagues: %v", err)
	} else {
		log.Println("performing league details parsing")
		for _, league := range *leagues {
			dl.LoadLeagueDetails <- league.ID
		}
	}

	return nil
}

// storeLeagues gets data from api and stores it into DB
func (dl *DataLoader) storeLeagues() error {
	leagueData, err := api.LoadLeagues()
	if err != nil {
		return err
	}

	hasRecord, err := dl.LeagueRepository.HasAnyRecord()
	if err != nil {
		return err
	}

	if hasRecord {
		for _, league := range leagueData.Leagues {
			//Store league only if it not exists in the DB
			b, _ := dl.LeagueRepository.ExistsByID(league.ID)

			if !b {
				if err = dl.LeagueRepository.Store(&league); err != nil {
					return err
				}
			}
		}

	} else {
		err = dl.LeagueRepository.StoreAll(&leagueData.Leagues)
		if err != nil {
			return err
		}
	}

	return nil
}

// storeLeagueDetails gets data from api and stores it into DB (upsert + series refresh).
// Records refreshed within detailsRefreshInterval are skipped to spare the Valve API.
func (dl *DataLoader) storeLeagueDetails(leagueID int) error {
	exist, err := dl.LeagueDetailsRepository.ExistsByID(leagueID)
	if err != nil {
		return err
	}

	if exist {
		stored, gerr := dl.LeagueDetailsRepository.GetByID(leagueID)
		switch {
		case gerr != nil:
			log.Printf("storeLeagueDetails: GetByID(%d) error: %s", leagueID, gerr)
		case time.Since(time.Unix(stored.UpdatedTimestamp, 0)) < detailsRefreshInterval:
			return nil
		}
	}

	leagueDetails, err := api.LoadLeagueDetails(leagueID)
	if err != nil {
		return err
	}

	// add pricepool info to details struct
	leagueDetails.Details.BasePrizePool = leagueDetails.PrizePool.BasePrizePool
	leagueDetails.Details.TotalPrizePool = leagueDetails.PrizePool.TotalPrizePool

	// add stream info
	leagueDetails.Details.Streams = leagueDetails.Streams

	if err := dl.LeagueSeriesRepository.ReplaceAllForLeague(leagueID, leagueDetails.SeriesInfos); err != nil {
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

			exists, err := dl.TeamRepository.ExistsByID(teamID)
			if err != nil {
				// transient DB failure is not a missing team - do not fan out an api request for it
				log.Printf("storeLeagueDetails: ExistsByID(%d) error: %s", teamID, err)
				continue
			}

			if !exists {
				dl.LoadTeam <- teamID
			}
		}
	}

	leagueDetails.Details.UpdatedTimestamp = time.Now().Unix()
	if exist {
		err = dl.LeagueDetailsRepository.Update(&leagueDetails.Details)
	} else {
		err = dl.LeagueDetailsRepository.Store(&leagueDetails.Details)
	}
	if err != nil {
		return err
	}

	err = dl.downloadLeagueImage(leagueID)
	if err != nil {
		log.Printf("download image error %s", err)
	}

	return nil
}
