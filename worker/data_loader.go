package worker

import (
	"dota_league/repository"
	"log"
	"time"
)

// DataLoader struct
type DataLoader struct {
	LeagueRepository        *repository.LeagueRepositoryInterface
	LeagueDetailsRepository *repository.LeagueDetailsRepositoryInterface
	GameRepository          *repository.GameRepositoryInterface
	LoadLeagueDetails       chan int
	LeaguesTicker           *time.Ticker
	GamesTicker             *time.Ticker
}

// NewDataLoader - create DataLoader and run worker
func NewDataLoader(
	lr *repository.LeagueRepositoryInterface,
	ldr *repository.LeagueDetailsRepositoryInterface,
	gr *repository.GameRepositoryInterface) *DataLoader {

	dataLoader := &DataLoader{
		LeagueRepository:        lr,
		LeagueDetailsRepository: ldr,
		GameRepository:          gr,
		LoadLeagueDetails:       make(chan int),

		//First leagues tick 2 seconds after start
		LeaguesTicker: time.NewTicker(2 * time.Second),

		// update live games every minute
		GamesTicker: time.NewTicker(1 * time.Minute),
	}

	go dataLoader.run()

	return dataLoader
}

func (dl *DataLoader) run() {
	defer dl.stop()

	// after start we need to set all leagues as inactive
	(*dl.LeagueDetailsRepository).SetAllAsNotLive()

	n := 1
	for {
		select {
		case <-dl.LeaguesTicker.C:

			// next tick in 12 hours
			dl.LeaguesTicker = time.NewTicker(12 * time.Hour)

			go dl.performLeaguesUpdate()

		case <-dl.GamesTicker.C:
			go dl.performGamesUpdate()

		case leagueID := <-dl.LoadLeagueDetails:

			sleepTime := 1 * time.Second
			// perform each 10th request once in 3 sec to avoid ban
			if n >= 10 {
				sleepTime = 3 * time.Second
				n = 0
			}
			n++

			err := dl.storeLeagueDetails(leagueID, sleepTime)
			if err != nil {
				log.Printf("StoreLeagueDetails error: %s", err)
			}

		}
	}

}

func (dl *DataLoader) stop() {
	dl.LeaguesTicker.Stop()
	dl.GamesTicker.Stop()
	close(dl.LoadLeagueDetails)
}
