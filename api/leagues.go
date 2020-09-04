package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// LoadLeagues loads json from the Dota API
func LoadLeagues() (*model.LeagueData, error) {
	op := "api.LoadLeagues"
	body, err := doRequest("https://www.dota2.com/webapi/IDOTA2League/GetLeagueInfoList/v001?")
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer body.Close()

	leagueDataJSON := model.LeagueData{}
	err = json.NewDecoder(body).Decode(&leagueDataJSON)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &leagueDataJSON, nil
}

// LoadLeagueDetails loads json from the Dota API for the given leagueID
func LoadLeagueDetails(leagueID int) (*model.LeagueDetailsData, error) {
	op := "api.LoadLeagueDetails"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2League/GetLeagueData/v001?league_id=%d", leagueID)
	body, err := doRequest(url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer body.Close()

	leaguesDetailsJSON := model.LeagueDetailsData{}

	err = json.NewDecoder(body).Decode(&leaguesDetailsJSON)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &leaguesDetailsJSON, nil
}

// PrizePoolResponse api json struct
type PrizePoolResponse struct {
	PrizePool          int     `json:"prize_pool"`
	IncrementPerSecond float64 `json:"increment_per_second"`
}

// LoadPrizePool loads prize_pool Dota API for the given leagueID
func LoadPrizePool(leagueID int) (*PrizePoolResponse, error) {
	op := "api.LoadPrizePool"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2League/GetPrizePool/v001?league_id=%d", leagueID)
	body, err := doRequest(url)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	responseData, err := ioutil.ReadAll(body)

	if err != nil {
		log.Print("here")
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	//sometimes api returns null
	if string(responseData) == "null" {
		return nil, &e.Error{Code: e.EINTERNAL, Op: "api.LoadPrizePool - null response recieved"}
	}

	prizePoolResponse := PrizePoolResponse{}
	err = json.Unmarshal(responseData, &prizePoolResponse)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &prizePoolResponse, nil
}
