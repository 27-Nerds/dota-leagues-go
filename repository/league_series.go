package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"fmt"

	driver "github.com/arangodb/go-driver/v2/arangodb/shared"
)

// LeagueSeriesRepository struct
type LeagueSeriesRepository struct {
	Conn Database
}

// NewLeagueSeriesRepository create new repository for the team data in DB
func NewLeagueSeriesRepository(Conn Database) *LeagueSeriesRepository {

	repo := &LeagueSeriesRepository{Conn: Conn}
	return repo
}

// ReplaceAllForLeague replaces all series rows of the given league with the given slice
func (rs *LeagueSeriesRepository) ReplaceAllForLeague(ctx context.Context, leagueID int, series []model.SeriesInfo) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	rows := make([]model.LeagueSeries, 0, len(series))
	for _, s := range series {
		rows = append(rows, model.LeagueSeries{
			LeagueID:   leagueID,
			SeriesID:   s.SeriesID,
			SeriesType: s.SeriesType,
			StartTime:  s.StartTime,
			MatchIDs:   s.MatchIDs,
			TeamID1:    s.TeamID1,
			TeamID2:    s.TeamID2,
			DBKey:      fmt.Sprintf("%d_%d", leagueID, s.SeriesID),
		})
	}

	err := rs.Conn.WithTransaction(ctx, "league_series", func(txCtx context.Context) error {
		if err := rs.Conn.DoQueryBuilder(txCtx,
			"FOR s IN league_series FILTER s.league_id == @id REMOVE s IN league_series",
			map[string]any{"id": leagueID}); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return rs.Conn.InsertMany(txCtx, "league_series", rows)
	})
	if err != nil {
		return &e.Error{Op: "LeagueSeriesRepository.ReplaceAllForLeague", Err: err}
	}
	return nil
}

// GetAllByLeague returns series of the league with team names joined from teams collection
func (rs *LeagueSeriesRepository) GetAllByLeague(ctx context.Context, leagueID int, offset int, limit int) ([]model.LeagueSeries, int64, error) {
	// LIMIT runs before the joins so DOCUMENT() lookups happen for the page only (primary key hits)
	query := `FOR s IN league_series FILTER s.league_id == @id
SORT s.start_time DESC
LIMIT @offset, @limit
LET t1 = DOCUMENT("teams", TO_STRING(s.team_id_1))
LET t2 = DOCUMENT("teams", TO_STRING(s.team_id_2))
RETURN {league_id: s.league_id, series_id: s.series_id, series_type: s.series_type, start_time: s.start_time, match_ids: s.match_ids, team_id_1: s.team_id_1, team_id_2: s.team_id_2, team_name_1: t1.name, team_tag_1: t1.tag, team_name_2: t2.name, team_tag_2: t2.tag}`

	bindVars := map[string]any{
		"id":     leagueID,
		"offset": offset,
		"limit":  limit,
	}

	series := []model.LeagueSeries{}
	var totalCount int64

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	cursor, err := rs.Conn.QueryAll(ctx, query, bindVars, true)
	if e.IsNotFound(err) {
		return series, 0, nil
	} else if err != nil {
		return nil, 0, &e.Error{Op: "LeagueSeriesRepository.GetAllByLeague", Err: err}
	}
	defer db.CloseCursor(cursor)

	for {
		var doc model.LeagueSeries
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, 0, &e.Error{Op: "LeagueSeriesRepository.GetAllByLeague", Err: err}
		}
		series = append(series, doc)
	}

	totalCount = int64(cursor.Statistics().FullCountInt)

	return series, totalCount, nil
}

// GetSitemapMatches lists unique stored schedule URLs; it never fetches match details.
func (rs *LeagueSeriesRepository) GetSitemapMatches(ctx context.Context, offset, limit int) ([]model.MatchReference, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	query := `FOR s IN league_series
 FILTER s.league_id > 0
 FOR raw IN (IS_ARRAY(s.match_ids) ? s.match_ids : [])
 LET candidateID = TO_STRING(raw)
 FILTER REGEX_TEST(candidateID, "^[1-9][0-9]{0,19}$")
 COLLECT leagueID = s.league_id, matchID = candidateID
 SORT leagueID ASC, matchID ASC
 LIMIT @offset, @limit
 RETURN {league_id: leagueID, match_id: matchID}`
	cursor, err := rs.Conn.QueryAll(ctx, query, map[string]any{"offset": offset, "limit": limit}, true)
	if e.IsNotFound(err) {
		return []model.MatchReference{}, 0, nil
	}
	if err != nil {
		return nil, 0, &e.Error{Op: "LeagueSeriesRepository.GetSitemapMatches", Err: err}
	}
	defer db.CloseCursor(cursor)
	rows := []model.MatchReference{}
	for {
		var row model.MatchReference
		_, err := cursor.ReadDocument(ctx, &row)
		if driver.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, 0, &e.Error{Op: "LeagueSeriesRepository.GetSitemapMatches", Err: err}
		}
		rows = append(rows, row)
	}
	return rows, int64(cursor.Statistics().FullCountInt), nil
}
