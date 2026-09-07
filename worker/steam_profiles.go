package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"time"
)

// playerFeedName mirrors the page heading: professional name, then Steam name.
func playerFeedName(player *model.Player) string {
	if player.Name != "" {
		return player.Name
	}
	return player.SteamName
}

func (dl *DataLoader) performSteamProfilesUpdate(ctx context.Context) error {
	return dl.refreshSteamProfiles(ctx, api.LoadSteamPlayer)
}

func (dl *DataLoader) refreshSteamProfiles(ctx context.Context, load func(context.Context, int) (*model.Player, error)) error {
	ids, err := dl.PlayerRepository.GetSteamRefreshCandidates(ctx, time.Now(), 50)
	if err != nil {
		return err
	}
	// The Steam rate limiter inside the loader paces requests; no extra pause is needed.
	completed, unavailable := 0, 0
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		previous, err := dl.PlayerRepository.GetByID(ctx, id)
		if err != nil && !e.IsNotFound(err) {
			return err
		}
		profile, fetchErr := load(ctx, id)
		if err := ctx.Err(); err != nil {
			return err
		}
		status := "ok"
		if fetchErr != nil {
			profile = nil
			status = "unavailable"
			if !e.IsNotFound(fetchErr) {
				status = "request_failed"
			}
			unavailable++
		} else {
			completed++
		}
		if err := dl.PlayerRepository.SaveSteamEnrichment(ctx, id, profile, time.Now(), status); err != nil {
			return err
		}
		if previous != nil && status != "request_failed" {
			if current, err := dl.PlayerRepository.GetByID(ctx, id); err == nil && current != nil {
				dl.recordUpdate(ctx, "player", id, playerFeedName(current), steamSnapshot(previous), steamSnapshot(current))
			}
		}
		// A transport/rate-limit failure may affect Steam globally. Pause this sweep
		// instead of sending another 49 requests; retry metadata prevents starvation.
		if status == "request_failed" {
			return fetchErr
		}
	}
	if len(ids) > 0 {
		slog.InfoContext(ctx, "Steam profiles refreshed", "updated", completed, "unavailable", unavailable)
	}
	return nil
}
