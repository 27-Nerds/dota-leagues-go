// Package worker refreshes league, team, player, and live-match data.
package worker

import (
	"log"
	"time"
)

// channelBuffer absorbs short bursts; producers wait when consumers fall behind.
const channelBuffer = 1024

// detailsRefreshInterval - do not refetch a league/team that was refreshed within this interval
const detailsRefreshInterval = 24 * time.Hour

// DataLoader struct
type DataLoader struct {
	LeagueRepository          LeagueRepository
	LeagueDetailsRepository   LeagueDetailsRepository
	GameRepository            GameRepository
	PlayerRepository          PlayerRepository
	TeamRepository            TeamRepository
	TeamRosterRepository      TeamRosterRepository
	LeagueSeriesRepository    LeagueSeriesRepository
	LiveGameDetailsRepository LiveGameDetailsRepository
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
	lr LeagueRepository,
	ldr LeagueDetailsRepository,
	gr GameRepository,
	pr PlayerRepository,
	tr TeamRepository,
	trr TeamRosterRepository,
	lsr LeagueSeriesRepository,
	lgdr LiveGameDetailsRepository,
) *DataLoader {

	dataLoader := &DataLoader{
		LeagueRepository:          lr,
		LeagueDetailsRepository:   ldr,
		GameRepository:            gr,
		PlayerRepository:          pr,
		TeamRepository:            tr,
		TeamRosterRepository:      trr,
		LeagueSeriesRepository:    lsr,
		LiveGameDetailsRepository: lgdr,
		LoadLeagueDetails:         make(chan int, channelBuffer),
		LoadTeam:                  make(chan int, channelBuffer),
		LoadSinglePlayer:          make(chan int, channelBuffer),

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
	if err := dl.LeagueDetailsRepository.SetAllAsNotLive(); err != nil {
		log.Printf("reset league live status: %v", err)
	}

	for {
		select {
		case <-dl.LeaguesTicker.C:

			// next tick in 12 hours
			dl.LeaguesTicker = time.NewTicker(12 * time.Hour)

			go runUpdate("leagues", dl.performLeaguesUpdate)

		case <-dl.GamesTicker.C:
			go runUpdate("games", dl.performGamesUpdate)

		case <-dl.PrizePoolTicker.C:
			go runUpdate("prizepool", dl.performPrizePoolUpdate)

		case <-dl.PlayersTicker.C:

			// next players tick in 12 hours
			dl.PlayersTicker = time.NewTicker(12 * time.Hour)

			go runUpdate("players", dl.performPlayersUpdate)
		}

	}
}

func (dl *DataLoader) runLoadLeagueDetails() {
	for leagueID := range dl.LoadLeagueDetails {
		if err := dl.storeLeagueDetails(leagueID); err != nil {
			log.Printf("storeLeagueDetails error: %s", err)
		}
	}
}

func (dl *DataLoader) runLoadTeam() {
	for teamID := range dl.LoadTeam {
		if err := dl.storeTeam(teamID); err != nil {
			log.Printf("storeTeam error: %s", err)
		}
	}
}

func (dl *DataLoader) runLoadSinglePlayer() {
	for playerID := range dl.LoadSinglePlayer {
		if err := dl.storeSinglePlayer(playerID); err != nil {
			log.Printf("storeSinglePlayer error: %s", err)
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

// runUpdate reports errors from scheduled work at the goroutine boundary.
func runUpdate(name string, update func() error) {
	if err := update(); err != nil {
		log.Printf("%s update failed: %v", name, err)
	}
}
