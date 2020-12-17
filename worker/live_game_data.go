package worker

import (
	"dota_league/api"
	"dota_league/model"
	"dota_league/repository"
	"log"
	"strconv"
	"time"
)

const (
	timeout = 200 * time.Second
)

//LiveGamesManager manages structs for every live game
type LiveGamesManager struct {
	liveGames                 map[int64]*LiveGame
	liveGameDetailsRepository *repository.LiveGameDetailsInterface
	gameEndedChannel          chan int64
}

func NewLiveGamesManager(liveGameDetailsRepository *repository.LiveGameDetailsInterface) *LiveGamesManager {
	log.Println("Live Game Manager started")
	lgm := &LiveGamesManager{
		liveGames:                 make(map[int64]*LiveGame),
		liveGameDetailsRepository: liveGameDetailsRepository,
		gameEndedChannel:          make(chan int64),
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
				log.Printf("updateLiveGameData error for %d. %v", serverSteamID, err)
			} else {
				liveGame.NewDataChan <- liveGameDetails

				time.Sleep(2 * time.Second)
			}
		}

		time.Sleep(3 * time.Second)
	}

}

func (lgm *LiveGamesManager) gameEndedListener() {
	for {
		select {
		case gameID := <-lgm.gameEndedChannel:
			// delete game from the games map
			delete(lgm.liveGames, gameID)
		}
	}
}

//LiveGame struct
type LiveGame struct {
	game                      model.Game
	liveGameDetailsRepository *repository.LiveGameDetailsInterface
	timeoutTicker             *time.Ticker
	gameEndedChannel          chan int64
	NewDataChan               chan *model.LiveGameDetails
}

//NewLiveGame create new live game for given id
func NewLiveGame(game model.Game, liveGameDetailsRepository *repository.LiveGameDetailsInterface, gameEndedChannel chan int64) *LiveGame {
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

	lgd.DbKey = strconv.FormatInt(lgd.Match.Matchid, 10)
	exist, err := (*lg.liveGameDetailsRepository).ExistsByID(lgd.Match.Matchid)
	if err != nil {
		log.Printf("updateLiveGameData ExistsByID error for %d. %v", lg.game.ServerSteamID, err)

		return err
	}
	if exist != true {
		err = (*lg.liveGameDetailsRepository).Store(lgd)
		if err != nil {
			log.Printf("updateLiveGameData store error for %d. %v", lg.game.ServerSteamID, err)

			return err
		}
	} else {
		err = (*lg.liveGameDetailsRepository).Update(lgd)
		if err != nil {
			log.Printf("updateLiveGameData store error for %d. %v", lg.game.ServerSteamID, err)

			return err
		}
	}

	return nil
}
