package api

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// LoadPlayerWithSteamFallback uses Steam identity only when DPC explicitly has no profile.
func LoadPlayerWithSteamFallback(ctx context.Context, id int) (*model.Player, error) {
	return loadPlayerWithFallback(ctx, id, LoadSinglePlayer, LoadSteamPlayer)
}

func loadPlayerWithFallback(ctx context.Context, id int, dpc, steam func(context.Context, int) (*model.Player, error)) (*model.Player, error) {
	player, err := dpc(ctx, id)
	if e.IsNotFound(err) && ctx.Err() == nil {
		return steam(ctx, id)
	}
	return player, err
}

// LoadSteamPlayer reads the token-free Community XML endpoint under the separate Steam rate limit.
func LoadSteamPlayer(ctx context.Context, id int) (*model.Player, error) {
	const op = "api.LoadSteamPlayer"
	if id <= 0 || uint64(id) > 4294967295 {
		return nil, &e.Error{Op: op, Code: e.EINVALID, Message: "invalid account ID"}
	}
	steamID := strconv.FormatUint(76561197960265728+uint64(id), 10)
	body, err := doSteamRequest(ctx, "https://steamcommunity.com/profiles/"+steamID+"/?xml=1")
	if err != nil {
		return nil, err
	}
	defer closeResponse(body)
	return decodeSteamPlayer(io.LimitReader(body, 1<<20), id, steamID)
}

func decodeSteamPlayer(body io.Reader, id int, steamID string) (*model.Player, error) {
	const op = "api.LoadSteamPlayer"
	var profile struct {
		XMLName  xml.Name
		Error    string `xml:"error"`
		ID       string `xml:"steamID64"`
		Name     string `xml:"steamID"`
		Avatar   string `xml:"avatarFull"`
		Location string `xml:"location"`
		Privacy  string `xml:"privacyState"`
	}
	if err := xml.NewDecoder(body).Decode(&profile); err != nil {
		return nil, &e.Error{Op: op, Code: e.EINVALID, Err: err}
	}
	if profile.XMLName.Local == "response" && profile.Error != "" && profile.ID == "" {
		return nil, &e.Error{Op: op, Code: e.ENOTFOUND, Message: "Steam profile unavailable"}
	}
	if profile.XMLName.Local != "profile" || profile.ID != steamID {
		return nil, &e.Error{Op: op, Code: e.EINVALID, Message: fmt.Sprintf("unexpected Steam profile for account %d", id)}
	}
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return nil, &e.Error{Op: op, Code: e.ENOTFOUND, Message: "Steam display name unavailable"}
	}
	avatar := ""
	if u, err := url.Parse(profile.Avatar); err == nil && u.Scheme == "https" && u.User == nil && (strings.HasSuffix(u.Hostname(), ".steamstatic.com") || strings.HasSuffix(u.Hostname(), ".steamcommunity.com")) {
		avatar = u.String()
	}
	privacy := strings.ToLower(strings.TrimSpace(profile.Privacy))
	if privacy == "" {
		privacy = "public"
	}
	return &model.Player{SteamName: name, SteamPrivacy: privacy, SteamUpdatedAt: time.Now().UnixMilli(), SteamNextRefreshAt: time.Now().Add(24 * time.Hour).UnixMilli(), SteamStatus: "ok", ID: id, Name: name, AvatarURL: avatar, SteamLocation: strings.TrimSpace(profile.Location), ProfileSource: "steam", ProfileCheckedAt: time.Now().UnixMilli()}, nil
}
