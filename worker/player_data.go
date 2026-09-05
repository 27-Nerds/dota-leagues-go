package worker

import (
	"dota_league/api"
	"dota_league/model"
	"log"
)

func (dl *DataLoader) performPlayersUpdate() error {
	players, err := dl.storePlayers()
	if err != nil {
		log.Printf("error while storeLeagues: %v", err)
		return err
	}
	log.Println("players info updated.")
	seen := make(map[int]bool)
	for _, player := range *players {
		// skip players with team id = 0
		if player.TeamID == 0 || seen[player.TeamID] {
			continue
		}

		seen[player.TeamID] = true
		dl.LoadTeam <- player.TeamID
	}

	return nil
}

// storePlayers gets data from api and stores it into DB
func (dl *DataLoader) storePlayers() (*[]model.Player, error) {
	playersData, err := api.LoadPlayers()
	if err != nil {
		return nil, err
	}

	hasRecord, err := dl.PlayerRepository.HasAnyRecord()
	if err != nil {
		return nil, err
	}

	if hasRecord {
		for _, player := range playersData.Players {
			//TODO: we need to update values sometimes

			//Store player only if it not exists in the DB
			b, _ := dl.PlayerRepository.ExistsByID(player.ID)

			if !b {
				if err = dl.PlayerRepository.Store(&player); err != nil {
					return nil, err
				}
			}
		}

	} else {
		err = dl.PlayerRepository.StoreAll(&playersData.Players)
		if err != nil {
			return nil, err
		}
	}

	return &playersData.Players, nil
}

func (dl *DataLoader) storeSinglePlayer(playerID int) error {
	player, err := api.LoadSinglePlayer(playerID)
	if err != nil {
		return err
	}

	err = dl.PlayerRepository.Store(player)
	if err != nil {
		return err
	}

	return nil
}
