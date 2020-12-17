package worker

import (
	"dota_league/api"
	"log"
)

func (dl *DataLoader) performGamesUpdate() error {
	log.Println("refreshing live games...")

	liveGames, err := api.LoadLiveGames()
	if err != nil {
		log.Printf("performGamesUpdate LoadLiveGames error: %s", err)
		return err
	}

	// do not do anything if there are no new data
	if len(liveGames.Games) == 0 {
		log.Println("No New games fetched")
		return nil
	}

	previousGames, err := (*dl.GameRepository).GetAll()
	if err != nil {
		log.Printf("performGamesUpdate error: %s", err)
		return err
	}

	finishedLeagues := make(map[int]bool)
	activeLeagues := make(map[int]bool)

	//TODO maybe move it to db query

	// build the list of finished games
	if len(*previousGames) > 0 {
		for _, pGame := range *previousGames {

			found := false
			for _, lGame := range liveGames.Games {
				if lGame.LeagueID == pGame.LeagueID {
					found = true
					activeLeagues[lGame.LeagueID] = true
					break
				}
			}
			if found == false {
				finishedLeagues[pGame.LeagueID] = true
			}
		}
		log.Printf("finised leagues: %v", finishedLeagues)
		log.Printf("Active leagues: %v", activeLeagues)

		// add games to the manager
		for _, lGame := range liveGames.Games {
			dl.LiveGamesManager.AddGame(lGame)
		}
	}

	for activeLeagueID := range activeLeagues {
		leagueDetailsExist, err := (*dl.LeagueDetailsRepository).ExistsByID(activeLeagueID)
		if err != nil {
			log.Printf("ExistsByID error: %v", err)
		}
		if leagueDetailsExist {
			err = (*dl.LeagueDetailsRepository).UpdateLiveStatus(activeLeagueID, true)
			if err != nil {
				log.Printf("UpdateLiveStatus error: %v", err)
			}
		} else {
			// if there is no league in the db, add it
			dl.LoadLeagueDetails <- activeLeagueID
		}
	}

	for finishedLeagueID := range finishedLeagues {
		err = (*dl.LeagueDetailsRepository).UpdateLiveStatus(finishedLeagueID, false)
		if err != nil {
			log.Printf("UpdateLiveStatus error: %v", err)
		}
	}

	err = (*dl.GameRepository).RemoveAll()
	if err != nil {
		log.Printf("performGamesUpdate error: %s", err)
		return err
	}

	err = (*dl.GameRepository).StoreAll(&liveGames.Games)
	if err != nil {
		log.Printf("performGamesUpdate error: %s", err)
		return err
	}

	return nil
}
