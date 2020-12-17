package worker

import (
	"dota_league/repository"
	"log"
	"time"
)

// DataLoader struct
type DataLoader struct {
	LeagueRepository          *repository.LeagueRepositoryInterface
	LeagueDetailsRepository   *repository.LeagueDetailsRepositoryInterface
	GameRepository            *repository.GameRepositoryInterface
	PlayerRepository          *repository.PlayerRepositoryInterface
	TeamRepository            *repository.TeamRepositoryInterface
	TeamRosterRepository      *repository.TeamRosterRepositoryInterface
	LiveGameDetailsRepository *repository.LiveGameDetailsInterface
	LoadLeagueDetails         chan int
	LoadTeam                  chan int
	LoadSinglePlayer          chan int
	LeaguesTicker             *time.Ticker
	GamesTicker               *time.Ticker
	PrizePoolTicker           *time.Ticker
	PlayersTicker             *time.Ticker
	LiveGamesManager          *LiveGamesManager
}

// NewDataLoader - create DataLoader and run worker
func NewDataLoader(
	lr *repository.LeagueRepositoryInterface,
	ldr *repository.LeagueDetailsRepositoryInterface,
	gr *repository.GameRepositoryInterface,
	pr *repository.PlayerRepositoryInterface,
	tr *repository.TeamRepositoryInterface,
	trr *repository.TeamRosterRepositoryInterface,
	lgdr *repository.LiveGameDetailsInterface,
) *DataLoader {

	dataLoader := &DataLoader{
		LeagueRepository:          lr,
		LeagueDetailsRepository:   ldr,
		GameRepository:            gr,
		PlayerRepository:          pr,
		TeamRepository:            tr,
		TeamRosterRepository:      trr,
		LiveGameDetailsRepository: lgdr,
		LoadLeagueDetails:         make(chan int),
		LoadTeam:                  make(chan int),
		LoadSinglePlayer:          make(chan int),

		//First leagues tick 2 seconds after start
		LeaguesTicker: time.NewTicker(2 * time.Second),

		// update live games every minute
		GamesTicker: time.NewTicker(1 * time.Minute),

		// update prizepool every hour
		PrizePoolTicker: time.NewTicker(1 * time.Hour),

		//First players tick 10 seconds after start
		PlayersTicker: time.NewTicker(10 * time.Second),
	}
	dataLoader.LiveGamesManager = NewLiveGamesManager(dataLoader.LiveGameDetailsRepository)

	go dataLoader.run()
	go dataLoader.runLoadLeagueDetails()
	go dataLoader.runLoadSinglePlayer()
	go dataLoader.runLoadTeam()

	return dataLoader
}

func (dl *DataLoader) run() {
	defer dl.stop()

	// after start we need to set all leagues as inactive
	(*dl.LeagueDetailsRepository).SetAllAsNotLive()

	for {
		select {
		case <-dl.LeaguesTicker.C:

			// next tick in 12 hours
			dl.LeaguesTicker = time.NewTicker(12 * time.Hour)

			go dl.performLeaguesUpdate()

		case <-dl.GamesTicker.C:
			go dl.performGamesUpdate()

		case <-dl.PrizePoolTicker.C:
			go dl.performPrizePoolUpdate()

		case <-dl.PlayersTicker.C:

			// next players tick in 12 hours
			dl.PlayersTicker = time.NewTicker(12 * time.Hour)

			go dl.performPlayersUpdate()
		}

	}
}

func (dl *DataLoader) runLoadLeagueDetails() {
	ln := 1

	for {
		select {
		case leagueID := <-dl.LoadLeagueDetails:

			sleepTime := 1 * time.Second
			// perform each 10th request once in 3 sec to avoid ban
			if ln >= 10 {
				sleepTime = 3 * time.Second
				ln = 0
			}
			ln++

			err := dl.storeLeagueDetails(leagueID, sleepTime)
			if err != nil {
				log.Printf("StoreLeagueDetails error: %s", err)
			}

		}
	}
}

func (dl *DataLoader) runLoadTeam() {
	tn := 1
	for {
		select {
		case teamID := <-dl.LoadTeam:

			sleepTime := 1 * time.Second
			// perform each 10th request once in 3 sec to avoid ban
			if tn >= 10 {
				sleepTime = 3 * time.Second
				tn = 0
			}
			tn++

			err := dl.storeTeam(teamID, sleepTime)
			if err != nil {
				log.Printf("StoreTeam error: %s", err)
			}
		}
	}
}

func (dl *DataLoader) runLoadSinglePlayer() {
	pn := 1
	for {
		select {
		case PlayerID := <-dl.LoadSinglePlayer:
			sleepTime := 1 * time.Second
			// perform each 10th request once in 3 sec to avoid ban
			if pn >= 10 {
				sleepTime = 3 * time.Second
				pn = 0
			}
			pn++

			err := dl.storeSinglePlayer(PlayerID, sleepTime)
			if err != nil {
				log.Printf("storeSinglePlayer error: %s", err)
			}
		}
	}
}

func (dl *DataLoader) stop() {
	dl.LeaguesTicker.Stop()
	dl.GamesTicker.Stop()
	dl.PrizePoolTicker.Stop()
	dl.PlayersTicker.Stop()

	close(dl.LoadLeagueDetails)
	close(dl.LoadTeam)
	close(dl.LoadSinglePlayer)
}
