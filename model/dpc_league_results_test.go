package model

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDPCLeagueResultsDecodesFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/league_results_dpc.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var res DPCLeagueResults
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(res.Results) == 0 || res.Results[0].Standing != 1 || res.Results[0].TeamID == 0 {
		t.Errorf("unexpected results decode: %+v", res.Results)
	}
	if len(res.Points) == 0 || len(res.Dollars) == 0 {
		t.Errorf("points/dollars should decode, got %d/%d", len(res.Points), len(res.Dollars))
	}
}
