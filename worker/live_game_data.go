package worker

import (
	"dota_league/api"
	"dota_league/model"
	"log"
	"time"
)

const (
	timeout = 200 * time.Second
)

// LiveGamesManager manages structs for every live game
type LiveGamesManager struct {
	liveGames                 map[string]*LiveGame
	liveGameDetailsRepository LiveGameDetailsRepository
	gameEndedChannel          chan string
}

func NewLiveGamesManager(liveGameDetailsRepository LiveGameDetailsRepository) *LiveGamesManager {
	log.Println("Live Game Manager started")
	lgm := &LiveGamesManager{
		liveGames:                 make(map[string]*LiveGame),
		liveGameDetailsRepository: liveGameDetailsRepository,
		gameEndedChannel:          make(chan string),
	}
	go lgm.updateGames()
	go lgm.gameEndedListener()

	return lgm
}

func (lgm *LiveGamesManager) AddGame(game model.Game) {
	_, ex := lgm.liveGames[game.ServerSteamID]
	// create new game if it not exists
	if !ex {
		lgm.liveGames[game.ServerSteamID] = NewLiveGame(game, lgm.liveGameDetailsRepository, lgm.gameEndedChannel)
	}
}

func (lgm *LiveGamesManager) updateGames() {
	for {
		for serverSteamID, liveGame := range lgm.liveGames {
			liveGameDetails, err := api.GetLiveGameStats(serverSteamID)
			if err != nil {
				log.Printf("updateLiveGameData error for %s. %v", serverSteamID, err)
			} else {
				liveGame.NewDataChan <- liveGameDetails

				time.Sleep(2 * time.Second)
			}
		}

		time.Sleep(3 * time.Second)
	}

}

func (lgm *LiveGamesManager) gameEndedListener() {
	for gameID := range lgm.gameEndedChannel {
		delete(lgm.liveGames, gameID)
	}
}

// LiveGame struct
type LiveGame struct {
	game                      model.Game
	liveGameDetailsRepository LiveGameDetailsRepository
	timeoutTicker             *time.Ticker
	gameEndedChannel          chan string
	NewDataChan               chan *model.LiveGameDetails
}

// NewLiveGame create new live game for given id
func NewLiveGame(game model.Game, liveGameDetailsRepository LiveGameDetailsRepository, gameEndedChannel chan string) *LiveGame {
	log.Println("Adding new live game:", game.ServerSteamID)

	liveGame := &LiveGame{
		game:                      game,
		liveGameDetailsRepository: liveGameDetailsRepository,
		timeoutTicker:             time.NewTicker(timeout),
		gameEndedChannel:          gameEndedChannel,
		NewDataChan:               make(chan *model.LiveGameDetails),
	}
	go liveGame.run()

	return liveGame
}

func (lg *LiveGame) run() {
	defer lg.stopGame()

	for {
		select {
		case lgd := <-lg.NewDataChan:
			//update
			err := lg.update(lgd)
			if err == nil {
				//game update was successful, set new timeout for a game
				lg.timeoutTicker = time.NewTicker(timeout)
			}
		case <-lg.timeoutTicker.C:
			//close in case of the timeout
			log.Println("timeout for game:", lg.game.ServerSteamID)
			return
		}
	}
}

func (lg *LiveGame) stopGame() {
	log.Println(lg.game.ServerSteamID, "ended")
	lg.gameEndedChannel <- lg.game.ServerSteamID
}

func (lg *LiveGame) update(lgd *model.LiveGameDetails) error {

	lgd.DBKey = lgd.Match.Matchid
	exist, err := lg.liveGameDetailsRepository.ExistsByID(lgd.Match.Matchid)
	if err != nil {
		log.Printf("updateLiveGameData ExistsByID error for %s. %v", lg.game.ServerSteamID, err)

		return err
	}
	if !exist {
		err = lg.liveGameDetailsRepository.Store(lgd)
		if err != nil {
			log.Printf("updateLiveGameData store error for %s. %v", lg.game.ServerSteamID, err)

			return err
		}
	} else {
		err = lg.liveGameDetailsRepository.Update(lgd)
		if err != nil {
			log.Printf("updateLiveGameData store error for %s. %v", lg.game.ServerSteamID, err)

			return err
		}
	}

	return nil
}
