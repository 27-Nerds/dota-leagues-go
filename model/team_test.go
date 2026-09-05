package model

import (
	"encoding/json"
	"os"
	"testing"
)

func TestTeamDecodesFullDpcFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/team_dpc_full.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var team Team
	if err := json.Unmarshal(data, &team); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if team.ID == 0 || team.Name == "" {
		t.Errorf("expected team id and name, got %+v", team)
	}
	if len(team.Members) == 0 {
		t.Error("expected members")
	} else {
		if team.Members[0].RealName != "Erik Engel" {
			t.Errorf("member real_name lost during decoding: %q", team.Members[0].RealName)
		}
		hasProName := false
		for _, m := range team.Members {
			if m.ProName != "" {
				hasProName = true
			}
		}
		if !hasProName {
			t.Error("expected at least one member with pro_name")
		}
	}
	if len(team.DpcResults) == 0 {
		t.Error("expected dpc_results")
	} else if team.DpcResults[0].LeagueID == 0 {
		t.Error("dpc result league_id missing")
	}
	if len(team.MemberStats) == 0 {
		t.Error("expected member_stats")
	}
	if len(team.TeamStats.PlayedHeroes) == 0 {
		t.Error("expected team_stats.played_heroes")
	} else if team.TeamStats.PlayedHeroes[0].HeroID == 0 {
		t.Error("played hero id missing")
	}
}

func TestTeamDecodesMinimalFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/team_dpc_minimal.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var team Team
	if err := json.Unmarshal(data, &team); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if team.ID == 0 || team.Name == "" {
		t.Errorf("expected team id and name, got %+v", team)
	}
	if len(team.DpcResults) > 0 {
		t.Errorf("minimal fixture should have no dpc_results, got %d", len(team.DpcResults))
	}
}
