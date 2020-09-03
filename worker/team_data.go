package worker

import (
	"dota_league/api"
	"dota_league/model"
	"fmt"
	"log"
	"time"
)

// storeTeam gets data from api and stores it into DB
func (dl *DataLoader) storeTeam(teamID int, sleepTime time.Duration) error {

	//TODO: we need to update team info
	// do not load data for team that are in the db
	exist, err := (*dl.TeamRepository).ExistsByID(teamID)
	if !exist && err == nil {

		//wait before api request to avoid ban
		time.Sleep(sleepTime)

		team, err := api.LoadTeamDetails(teamID)
		if err != nil {
			return err
		}

		err = (*dl.TeamRepository).Store(team)
		if err != nil {
			return err
		}

		err = dl.downloadTeamImage(team)
		if err != nil {
			log.Printf("download image error %s", err)
		}

	} else if err != nil {
		return err
	}

	return nil
}

func (dl *DataLoader) downloadTeamImage(team *model.Team) error {

	// skip if logo url is empty
	if team.URLLogo == "" {
		return nil
	}

	path := fmt.Sprintf("public/teams/%d", team.ID)

	err := api.DownloadImageIfNotExist(team.URLLogo, path, "logo.png")
	if err != nil {
		return err
	}

	return nil
}
