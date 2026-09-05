package model

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDPCStandingsDecodesFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/dpc_standings.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var dpc DPCStandings
	if err := json.Unmarshal(data, &dpc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(dpc.Standings) != 0 || len(dpc.Results) != 0 ||
		len(dpc.MajorWildcardStandings) != 0 || len(dpc.MajorGroupStandings) != 0 ||
		len(dpc.MajorPlayoffStandings) != 0 {
		t.Errorf("fixture is the empty-between-seasons envelope, got %+v", dpc)
	}
}

func TestDPCStandingsDecodesPopulatedSample(t *testing.T) {
	data := []byte(`{
		"results": [],
		"standings": [{"team_id": 1234, "points": 50}],
		"major_wildcard_standings": [],
		"major_group_standings": [["a", "b"]],
		"major_playoff_standings": []
	}`)

	var dpc DPCStandings
	if err := json.Unmarshal(data, &dpc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(dpc.Standings) != 1 || dpc.Standings[0].(map[string]any)["team_id"] != float64(1234) {
		t.Errorf("standings item should pass through as raw map, got %+v", dpc.Standings)
	}
}
