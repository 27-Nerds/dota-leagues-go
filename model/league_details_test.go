package model

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLeagueDetailsDecodesSeriesFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/league_data_series.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var details LeagueDetailsData
	if err := json.Unmarshal(data, &details); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if details.Details.ID == 0 || details.Details.Name == "" {
		t.Errorf("expected league info, got %+v", details.Details)
	}
	if len(details.SeriesInfos) == 0 {
		t.Fatal("expected series_infos")
	}

	s := details.SeriesInfos[0]
	if s.SeriesID == 0 {
		t.Error("series_id missing")
	}
	if len(s.MatchIDs) == 0 || s.MatchIDs[0] == "" {
		t.Error("match_ids should decode as non-empty strings")
	}
	if s.TeamID1 == 0 && s.TeamID2 == 0 {
		t.Error("expected at least one team id in series info")
	}

	if len(details.NodeGroups) == 0 {
		t.Fatal("expected node_groups")
	}
	n := details.NodeGroups[0]
	if n.NodeGroupID == 0 {
		t.Error("node_group_id missing")
	}
	if len(n.TeamStandings) > 0 && n.TeamStandings[0].TeamName == "" {
		t.Errorf("team standing names should decode, got %+v", n.TeamStandings[0])
	}
}

func TestLeagueDetailsDecodesEmptySeries(t *testing.T) {
	data, err := os.ReadFile("../testdata/league_data_sample.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var details LeagueDetailsData
	if err := json.Unmarshal(data, &details); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(details.SeriesInfos) > 0 {
		t.Errorf("sample league has no series_infos, got %d", len(details.SeriesInfos))
	}
}
