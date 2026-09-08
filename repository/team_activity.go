package repository

import (
	"context"
	"log/slog"
	"time"
)

// Refresh competition history after this interval. Requests keep using the last
// successful snapshot during refreshes or outages. Live games are never cached.
const teamActivityTTL = 30 * time.Second

type teamActivity struct {
	LastPlayed int64 `json:"lastPlayed"`
	Tier       int   `json:"tier"`
}

const teamActivityQuery = `RETURN MERGE(
    FOR s IN league_series
        FILTER s.start_time > 0 && s.start_time <= @now && LENGTH(s.match_ids) > 0
        LET league = DOCUMENT("league_details", TO_STRING(s.league_id))
        LET recentTier = s.start_time >= @recentSince ? TO_NUMBER(league.tier) : 0
        FOR id IN [s.team_id_1, s.team_id_2]
            FILTER id > 0
            COLLECT teamID = id AGGREGATE lastPlayed = MAX(s.start_time), tier = MAX(recentTier)
            RETURN { [TO_STRING(teamID)]: { lastPlayed, tier } }
)`

// WarmCompetitiveActivity prepares the ranking before HTTP starts accepting requests.
func (tr *TeamRepository) WarmCompetitiveActivity(ctx context.Context) error {
	_, err := tr.competitiveActivity(ctx, time.Now())
	return err
}

// Cold callers share a refresh; warm callers never wait for an expired snapshot.
func (tr *TeamRepository) competitiveActivity(ctx context.Context, now time.Time) (map[string]teamActivity, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tr.activityMu.Lock()
		cached := tr.activity
		if cached != nil && now.Before(tr.activityUntil) {
			tr.activityMu.Unlock()
			return cached, nil
		}
		if pending := tr.activityLoading; pending != nil {
			tr.activityMu.Unlock()
			if cached != nil {
				return cached, nil
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-pending:
				continue
			}
		}
		pending := make(chan struct{})
		tr.activityLoading = pending
		tr.activityMu.Unlock()
		if cached != nil {
			// A browser disconnect must not cancel a refresh shared by other visitors.
			go func() { _, _ = tr.refreshActivity(context.WithoutCancel(ctx), now, pending) }()
			return cached, nil
		}
		return tr.refreshActivity(ctx, now, pending)
	}
}

func (tr *TeamRepository) refreshActivity(ctx context.Context, now time.Time, pending chan struct{}) (map[string]teamActivity, error) {
	activity := map[string]teamActivity{}
	queryCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := tr.Conn.Query(queryCtx, teamActivityQuery, map[string]any{
		"now": now.Unix(), "recentSince": now.Add(-90 * 24 * time.Hour).Unix(),
	}, &activity)
	tr.activityMu.Lock()
	if err == nil {
		tr.activity = activity
	}
	// Back off after failures too, retaining the last successful snapshot.
	tr.activityUntil = time.Now().Add(teamActivityTTL)
	tr.activityLoading = nil
	close(pending)
	tr.activityMu.Unlock()
	if err != nil {
		slog.WarnContext(ctx, "refresh team competitive activity", "error", err)
	}
	return activity, err
}
