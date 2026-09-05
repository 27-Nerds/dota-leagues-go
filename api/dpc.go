// Package api retrieves league, team, player, and match data from Valve.
package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
)

// LoadDPCStandings loads DPC standings from the Dota API.
// Envelope arrays are season-dependent and may be empty between seasons,
// so items are kept as raw maps until a populated sample pins their shape.
func LoadDPCStandings() (*model.DPCStandings, error) {
	const op = "api.LoadDPCStandings"

	url := "https://www.dota2.com/webapi/IDOTA2DPC/GetDPCStandings/v001"
	body, err := doRequest(url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer closeResponse(body)

	var dpc model.DPCStandings
	if err = json.NewDecoder(body).Decode(&dpc); err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &dpc, nil
}
