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
	leagues, err := (*dl.LeagueRepository).GetAllActive()
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

	hasRecord, err := (*dl.LeagueRepository).HasAnyRecord()
	if err != nil {
		return err
	}

	if hasRecord {
		for _, league := range leagueData.Leagues {
			//Store league only if it not exists in the DB
			b, _ := (*dl.LeagueRepository).ExistsByID(league.ID)

			if !b {
				if err = (*dl.LeagueRepository).Store(&league); err != nil {
					return err
				}
			}
		}

	} else {
		err = (*dl.LeagueRepository).StoreAll(&leagueData.Leagues)
		if err != nil {
			return err
		}
	}

	return nil
}

// storeLeagueDetails gets data from api and stores it into DB
func (dl *DataLoader) storeLeagueDetails(leagueID int, sleepTime time.Duration) error {
	// do not load data for leagues that are in the db
	exist, err := (*dl.LeagueDetailsRepository).ExistsByID(leagueID)
	if exist && err == nil {

		//try to redownload image to existing league if file does not exist
		err = dl.downloadLeagueImage(leagueID)
		if err != nil {
			log.Printf("download image error %s", err)
		}
	} else if !exist && err == nil {

		//wait before api request to avoid ban
		time.Sleep(sleepTime)

		leagueDetails, err := api.LoadLeagueDetails(leagueID)
		if err != nil {
			return err
		}

		// add pricepool info to details struct
		leagueDetails.Details.BasePrizePool = leagueDetails.PrizePool.BasePrizePool
		leagueDetails.Details.TotalPrizePool = leagueDetails.PrizePool.TotalPrizePool

		err = (*dl.LeagueDetailsRepository).Store(&leagueDetails.Details)
		if err != nil {
			return err
		}

		err = dl.downloadLeagueImage(leagueID)
		if err != nil {
			log.Printf("download image error %s", err)
		}

	} else if err != nil {
		return err
	}
	return nil
}
