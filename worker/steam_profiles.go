package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"time"
)

func (dl *DataLoader) performSteamProfilesUpdate(ctx context.Context) error {
	return dl.refreshSteamProfiles(ctx, api.LoadSteamPlayer)
}

func (dl *DataLoader) refreshSteamProfiles(ctx context.Context, load func(context.Context, int) (*model.Player, error)) error {
	ids, err := dl.PlayerRepository.GetSteamRefreshCandidates(ctx, time.Now(), 50)
	if err != nil {
		return err
	}
	completed, unavailable := 0, 0
	for i, id := range ids {
		if i > 0 {
			timer := time.NewTimer(5 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
		if err := ctx.Err(); err != nil {
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
