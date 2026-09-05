package model

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMatchMinimalDecodesFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/match_minimal_dpc.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var mm MatchMinimal
	if err := json.Unmarshal(data, &mm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if mm.MatchID == "" || len(mm.Players) != 10 {
		t.Errorf("unexpected match decode: id=%q players=%d", mm.MatchID, len(mm.Players))
	}
	if mm.Tourney.RadiantTeamName == "" || mm.Tourney.DireTeamID == 0 {
		t.Errorf("tourney block should decode: %+v", mm.Tourney)
	}
	for i, p := range mm.Players {
		if p.AccountID == 0 || p.HeroID == 0 || len(p.Items) != 6 {
			t.Errorf("player %d should fully decode: %+v", i, p)
		}
	}
}
