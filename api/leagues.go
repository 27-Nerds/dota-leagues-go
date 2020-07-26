package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// LoadLeagues loads json from the Dota API
func LoadLeagues() (*model.LeagueData, error) {
	body, err := doRequest("https://www.dota2.com/webapi/IDOTA2League/GetLeagueInfoList/v001?")
	if err != nil {
		return nil, err
	}
	leagueDataJSON := model.LeagueData{}
	err = json.Unmarshal(body, &leagueDataJSON)
	if err != nil {
		return nil, err
	}

	return &leagueDataJSON, nil
}

// LoadLeagueDetails loads json from the Dota API for the given leagueID
func LoadLeagueDetails(leagueID int) (*model.LeagueDetailsData, error) {
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2League/GetLeagueData/v001?league_id=%d", leagueID)
	body, err := doRequest(url)
	if err != nil {
		return nil, err
	}

	leaguesDetailsJSON := model.LeagueDetailsData{}

	err = json.Unmarshal(body, &leaguesDetailsJSON)
	if err != nil {
		return nil, err
	}

	return &leaguesDetailsJSON, nil
}

// DownloadImageIfNotExist Download image
func DownloadImageIfNotExist(sourceURL string, relativeDestPath string, imageName string) error {
	op := "api.DownloadImageIfNotExist"
	basePath, err := os.Getwd()
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	path := fmt.Sprintf("%s/%s", basePath, relativeDestPath)

	fullPathWithName := fmt.Sprintf("%s/%s", path, imageName)

	_, err = os.Stat(fullPathWithName)
	if os.IsNotExist(err) {
		log.Printf("downloading image, %s", sourceURL)

		resp, err := http.Get(sourceURL)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}
		if resp.StatusCode != 200 {
			return &e.Error{Code: e.ENOTFOUND, Op: op}
		}
		defer resp.Body.Close()

		os.MkdirAll(path, os.ModePerm)
		out, err := os.Create(fullPathWithName)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

	}

	return nil
}
