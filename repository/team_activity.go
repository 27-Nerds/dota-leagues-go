package repository

import (
	"context"
	"time"
)

// Competition history may lag by at most this interval. Live games, team
// profiles and roster changes are deliberately not part of this cache.
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

// A single refresh serves concurrent requests. Waiters can cancel independently;
// a failed/canceled refresh is never cached and releases them to retry.
func (tr *TeamRepository) competitiveActivity(ctx context.Context, now time.Time) (map[string]teamActivity, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tr.activityMu.Lock()
		if tr.activity != nil && now.Before(tr.activityUntil) {
			activity := tr.activity
			tr.activityMu.Unlock()
			return activity, nil
		}
		if pending := tr.activityLoading; pending != nil {
			tr.activityMu.Unlock()
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

		activity := map[string]teamActivity{}
		queryCtx, cancel := context.WithTimeout(ctx, dbTimeout)
		_, err := tr.Conn.Query(queryCtx, teamActivityQuery, map[string]any{
			"now": now.Unix(), "recentSince": now.Add(-90 * 24 * time.Hour).Unix(),
		}, &activity)
		cancel()

		tr.activityMu.Lock()
		if err == nil {
			tr.activity = activity
			// Expire relative to the snapshot time, not completion of a slow query.
			tr.activityUntil = now.Add(teamActivityTTL)
		}
		tr.activityLoading = nil
		close(pending)
		tr.activityMu.Unlock()
		return activity, err
	}
}
