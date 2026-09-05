package model

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGameDecodesLiveGamesFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/live_games.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var live LiveGames
	if err := json.Unmarshal(data, &live); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(live.Games) == 0 {
		t.Fatal("expected at least one game")
	}

	g := live.Games[0]
	if g.ServerSteamID == "" {
		t.Error("server_steam_id should decode as non-empty string")
	}
	if g.MatchID == "" {
		t.Error("match_id should decode as non-empty string")
	}
	if g.LeagueID == 0 {
		t.Error("league_id missing")
	}
}
