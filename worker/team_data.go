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
			log.Printf("storeTeam: LoadTeamDetails error %s", err)
			return err
		}

		err = (*dl.TeamRepository).Store(team)
		if err != nil {
			log.Printf("storeTeam: store team error %s", err)
			return err
		}

		err = dl.downloadTeamImage(team)
		if err != nil {
			log.Printf("storeTeam: download image error %s", err)
		}

		err = dl.storeTeamRoster(team)
		if err != nil {
			log.Printf("storeTeam: storeTeamRoster error %s", err)

			return err
		}

	} else if err != nil {
		log.Printf("storeTeam: ExistsByID error %s", err)
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

func (dl *DataLoader) storeTeamRoster(team *model.Team) error {
	teamRoster := model.TeamRoster{
		TeamID: team.ID,
	}

	for _, playerID := range team.RegisteredMemberAccountIds {
		exist, err := (*dl.PlayerRepository).ExistsByID(playerID)
		if err != nil {
			return err
		}
		if exist != true {
			log.Printf("load team member with id: %d", playerID)
			dl.LoadSinglePlayer <- playerID
		}

		teamMember := model.TeamMember{
			AccountID: playerID,
			IsActive:  true,
		}
		teamRoster.TeamMembers = append(teamRoster.TeamMembers, teamMember)
	}

	err := (*dl.TeamRosterRepository).Store(&teamRoster)
	if err != nil {
		return err
	}

	return nil
}
