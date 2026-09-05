package api

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"strings"
	"testing"
)

func TestSteamProfileDecode(t *testing.T) {
	for _, tc := range []struct{ name, body, code string }{
		{"public", `<profile><steamID64>76561198253776481</steamID64><steamID><![CDATA[DOG & TIRED]]></steamID><avatarFull>https://avatars.fastly.steamstatic.com/test.jpg</avatarFull><location>Bangladesh</location><realname>ignored</realname></profile>`, ""},
		{"private with name", `<profile><steamID64>76561198253776481</steamID64><steamID>Private player</steamID><privacyState>private</privacyState></profile>`, ""},
		{"missing", `<response><error>Profile not found</error></response>`, e.ENOTFOUND},
		{"wrong account", `<profile><steamID64>76561198253776482</steamID64><steamID>Wrong</steamID></profile>`, e.EINVALID},
		{"HTML", `<html><body>Access denied</body></html>`, e.EINVALID},
		{"malformed", `<profile>`, e.EINVALID},
		{"empty name", `<profile><steamID64>76561198253776481</steamID64><steamID> </steamID></profile>`, e.ENOTFOUND},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := decodeSteamPlayer(strings.NewReader(tc.body), 293510753, "76561198253776481")
			if e.ErrorCode(err) != tc.code {
				t.Fatalf("got %v, want %s", err, tc.code)
			}
			if tc.code == "" && (p.ID != 293510753 || p.Name == "" || p.ProfileSource != "steam" || p.ProfileCheckedAt == 0 || p.CountryCode != "" || p.RealName != "" || p.IsPro || p.TeamID != 0) {
				t.Fatalf("unexpected identity: %+v", p)
			}
			if tc.name == "public" && (p.Name != "DOG & TIRED" || p.AvatarURL == "" || p.SteamLocation != "Bangladesh") {
				t.Fatalf("missing name/avatar: %+v", p)
			}
			if tc.name == "private with name" && p.SteamLocation != "" {
				t.Fatal("missing location should remain empty")
			}
		})
	}
}

func TestSteamFallbackOnlyForMissingDPC(t *testing.T) {
	for _, tc := range []struct {
		name      string
		err       error
		wantSteam bool
	}{
		{"DPC available", nil, false},
		{"DPC missing", &e.Error{Code: e.ENOTFOUND}, true},
		{"timeout", context.DeadlineExceeded, false},
		{"invalid response", &e.Error{Code: e.EINVALID}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			_, err := loadPlayerWithFallback(t.Context(), 42, func(context.Context, int) (*model.Player, error) { return &model.Player{ID: 42}, tc.err }, func(context.Context, int) (*model.Player, error) {
				called = true
				return &model.Player{ID: 42, ProfileSource: "steam"}, nil
			})
			if called != tc.wantSteam {
				t.Fatal("unexpected fallback call")
			}
			if !tc.wantSteam && err != tc.err {
				t.Fatal("original error lost")
			}
		})
	}
}

func TestSteamLocationPreservesTextAndOmitsEmptyValue(t *testing.T) {
	for _, location := range []string{"", "   ", "  Kyiv, Kyyivs'ka Oblast', Ukraine  "} {
		body := `<profile><steamID64>76561198253776481</steamID64><steamID>Player</steamID><location><![CDATA[` + location + `]]></location></profile>`
		p, err := decodeSteamPlayer(strings.NewReader(body), 293510753, "76561198253776481")
		if err != nil {
			t.Fatal(err)
		}
		if p.SteamLocation != strings.TrimSpace(location) || p.CountryCode != "" {
			t.Fatalf("unexpected location/nationality: %+v", p)
		}
		raw, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), `"steam_location"`) != (strings.TrimSpace(location) != "") {
			t.Fatal("empty location was not omitted")
		}
	}
}
