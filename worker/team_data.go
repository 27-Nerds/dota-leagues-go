package worker

import (
	"dota_league/api"
	"dota_league/model"
	"fmt"
	"log"
	"time"
)

// storeTeam gets data from api and stores or updates it in the DB.
// Teams refreshed within detailsRefreshInterval are skipped to spare the Valve API.
func (dl *DataLoader) storeTeam(teamID int) error {

	exist, err := dl.TeamRepository.ExistsByID(teamID)
	if err != nil {
		log.Printf("storeTeam: ExistsByID error %s", err)
		return err
	}

	if exist {
		stored, gerr := dl.TeamRepository.GetByID(teamID)
		switch {
		case gerr != nil:
			log.Printf("storeTeam: GetByID(%d) error: %s", teamID, gerr)
		case time.Since(time.Unix(stored.UpdatedTimestamp, 0)) < detailsRefreshInterval:
			return nil
		}
	}

	team, err := api.LoadTeamDetails(teamID)
	if err != nil {
		log.Printf("storeTeam: LoadTeamDetails error %s", err)
		return err
	}

	if err := dl.storeTeamRoster(team); err != nil {
		return err
	}

	team.UpdatedTimestamp = time.Now().Unix()

	if exist {
		err = dl.TeamRepository.Update(team)
	} else {
		err = dl.TeamRepository.Store(team)
	}
	if err != nil {
		log.Printf("storeTeam: store team error %s", err)
		return err
	}

	err = dl.downloadTeamImage(team)
	if err != nil {
		log.Printf("storeTeam: download image error %s", err)
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

	for _, member := range team.Members {
		// valve returns placeholder members (account_id 0) for empty roster slots
		if member.AccountID == 0 {
			continue
		}

		exist, err := dl.PlayerRepository.ExistsByID(member.AccountID)
		if err != nil {
			return err
		}
		if !exist {
			log.Printf("load team member with id: %d", member.AccountID)
			dl.LoadSinglePlayer <- member.AccountID
		}

		teamMember := model.TeamMember{
			AccountID: member.AccountID,
			IsActive:  true,
		}
		teamRoster.TeamMembers = append(teamRoster.TeamMembers, teamMember)
	}

	exist, err := dl.TeamRosterRepository.ExistsByTeamID(team.ID)
	if err != nil {
		return err
	}

	if exist {
		err = dl.TeamRosterRepository.Update(&teamRoster)
	} else {
		err = dl.TeamRosterRepository.Store(&teamRoster)
	}
	if err != nil {
		return err
	}

	return nil
}
