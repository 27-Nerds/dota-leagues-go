package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"strings"
	"sync"
	"time"

	driver "github.com/arangodb/go-driver/v2/arangodb/shared"
)

// TeamRepository repository object
type TeamRepository struct {
	Conn            Database
	activityMu      sync.Mutex
	activity        map[string]teamActivity
	activityUntil   time.Time
	activityLoading chan struct{}
}

// NewTeamRepository creates new struct
func NewTeamRepository(Conn Database) *TeamRepository {
	return &TeamRepository{Conn: Conn}
}

// Store store team model in db
func (tr *TeamRepository) Store(ctx context.Context, team *model.Team) error {
	team.DBKey = strconv.Itoa(team.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := tr.Conn.Insert(ctx, "teams", team)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "TeamRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "TeamRepository.Store", Err: err}
	}

	return nil
}

// ExistsByID check wether record exists in the DB
func (tr *TeamRepository) ExistsByID(ctx context.Context, id int) (bool, error) {

	exists, err := existsInColByID(ctx, tr.Conn, "teams", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "TeamRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// Update updates an existing team record (partial merge)
func (tr *TeamRepository) Update(ctx context.Context, team *model.Team) error {
	team.DBKey = strconv.Itoa(team.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := tr.Conn.Update(ctx, "teams", team.DBKey, team)
	if err != nil {
		return &e.Error{Op: "TeamRepository.Update", Err: err}
	}

	return nil
}

// GetByID get team by id
func (tr *TeamRepository) GetByID(ctx context.Context, id int) (*model.Team, error) {
	bindVars := map[string]any{
		"id": strconv.Itoa(id),
	}

	query := `FOR d IN teams FILTER d._key == @id
LET members = (
 FOR member IN (IS_ARRAY(d.members) ? d.members : [])
 LET player = DOCUMENT("players", TO_STRING(member.account_id))
 RETURN MERGE(member, {player_name: player.name, steam_location: player.steam_location, avatar_url: player.avatar_url})
)
LET history = (
 FOR result IN (IS_ARRAY(d.dpc_results) ? d.dpc_results : [])
 LET details = DOCUMENT("league_details", TO_STRING(result.league_id))
 LET league = DOCUMENT("leagues", TO_STRING(result.league_id))
 RETURN MERGE(result, {
  league_name: details.name != null && details.name != "" ? details.name : league.name,
  league_available: details != null && details.league_id > 0
 })
)
RETURN MERGE(d, {members, dpc_results: history})`

	var team model.Team
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := tr.Conn.Query(ctx, query, bindVars, &team)
	if err != nil {
		return nil, &e.Error{Op: "TeamRepository.GetByID", Err: err}
	}

	return &team, nil
}

// GetAll ranks teams by live play, recent tournament tier, and last match.
// Metadata refresh timestamps are deliberately excluded: refreshing an old team
// does not make it competitively active. ID breaks ties for stable pagination.
func (tr *TeamRepository) GetAll(ctx context.Context, offset int, limit int, filter model.TeamFilter) ([]model.Team, int64, error) {
	sort := "isLive DESC, isRecent DESC, stats.tier DESC, stats.lastPlayed DESC, d.team_id ASC"
	if field, ok := map[string]string{"name": "LOWER(d.name)", "wins": "TO_NUMBER(d.wins)", "activity": "lastActivity"}[filter.Sort]; ok {
		sort = field + " " + sortOrder(filter.Sort, filter.Order) + ", d.team_id ASC"
	}
	changes := "LET changes = {}"
	if filter.Active || filter.Sort == "activity" {
		changes = `LET changes = MERGE(
 FOR u IN updates
  FILTER u.entity IN ["team", "roster"] && u.action == "updated"
  FILTER u.created_at < (@now + 1) * 1000
  COLLECT teamID = u.entity_id AGGREGATE lastChanged = MAX(u.created_at)
  RETURN { [TO_STRING(teamID)]: lastChanged / 1000 }
)`
	}
	query := `
LET liveTeams = UNIQUE(FLATTEN(
    FOR g IN games RETURN [g.radiant_team_id, g.dire_team_id]
))
LET activity = @activity
` + changes + `
FOR d IN teams
    FILTER d.team_id > 0
    FILTER @search == "" || CONTAINS(LOWER(d.name), @search) || CONTAINS(LOWER(d.tag), @search)
    FILTER @country == "" || UPPER(d.country_code) == @country
    FILTER @pro == null || d.pro == @pro
    LET stats = activity[TO_STRING(d.team_id)]
    LET isLive = d.team_id IN liveTeams
    LET isRecent = stats.lastPlayed >= @recentSince
    LET lastActivity = isLive ? @now : MAX([TO_NUMBER(stats.lastPlayed), TO_NUMBER(changes[TO_STRING(d.team_id)])])
    FILTER !@activeOnly || lastActivity >= @activeSince
    SORT ` + sort + `
    LIMIT @offset, @limit
    RETURN KEEP(d, "team_id", "name", "tag", "country_code", "region", "pro", "wins")`
	now := time.Now()
	days := filter.ActiveDays
	if days < 1 || days > 3650 {
		days = 90
	}
	bindVars := map[string]any{
		"offset":      offset,
		"limit":       limit,
		"search":      strings.ToLower(strings.TrimSpace(filter.Search)),
		"country":     strings.ToUpper(strings.TrimSpace(filter.Country)),
		"activeOnly":  filter.Active,
		"pro":         filter.Pro,
		"activeSince": now.Add(-time.Duration(days) * 24 * time.Hour).Unix(),
		"now":         now.Unix(),
		"recentSince": now.Add(-90 * 24 * time.Hour).Unix(),
	}

	// Name/wins sorting without activity filtering needs no competition history.
	activity := map[string]teamActivity{}
	if filter.Active || (filter.Sort != "name" && filter.Sort != "wins") {
		var err error
		activity, err = tr.competitiveActivity(ctx, now)
		if err != nil {
			return nil, 0, &e.Error{Op: "TeamRepository.GetAll.activity", Err: err}
		}
	}
	bindVars["activity"] = activity

	teams := []model.Team{}
	var totalCount int64

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	cursor, err := tr.Conn.QueryAll(ctx, query, bindVars, true)
	if err != nil {
		return nil, 0, &e.Error{Op: "TeamRepository.GetAll", Err: err}
	}
	defer db.CloseCursor(cursor)

	for {
		var doc model.Team
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, 0, &e.Error{Op: "TeamRepository.GetAll", Err: err}
		}
		teams = append(teams, doc)
	}

	totalCount = int64(cursor.Statistics().FullCountInt)
	return teams, totalCount, nil
}

// GetRefreshCandidates prioritizes recent competition and roster activity.
// Old teams are still checked daily so disbands and comebacks remain observable.
func (tr *TeamRepository) GetRefreshCandidates(ctx context.Context, now time.Time) ([]int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	var ids []int
	_, err := tr.Conn.Query(ctx, `RETURN (
 LET competing = UNIQUE(FLATTEN(
  FOR s IN league_series
   FILTER s.start_time >= @recentSince && s.start_time <= @upcomingUntil
   RETURN [s.team_id_1, s.team_id_2]
 ))
 LET changing = UNIQUE(
  FOR u IN updates
   FILTER u.entity == "roster" && u.action == "updated"
   FILTER u.created_at >= @changedSince && u.created_at <= @now * 1000
   RETURN u.entity_id
 )
 LET live = UNIQUE(FLATTEN(FOR g IN games RETURN [g.radiant_team_id, g.dire_team_id]))
 FOR team IN teams
  FILTER team.team_id > 0
  LET active = team.team_id IN competing || team.team_id IN changing || team.team_id IN live
  FILTER TO_NUMBER(team.updated_timestamp) <= (active ? @activeBefore : @inactiveBefore)
  SORT active DESC, TO_NUMBER(team.updated_timestamp) ASC, team.team_id ASC
  RETURN team.team_id
 )`, map[string]any{
		"now":            now.Unix(),
		"recentSince":    now.Add(-90 * 24 * time.Hour).Unix(),
		"upcomingUntil":  now.Add(30 * 24 * time.Hour).Unix(),
		"changedSince":   now.Add(-30 * 24 * time.Hour).UnixMilli(),
		"activeBefore":   now.Add(-2 * time.Hour).Unix(),
		"inactiveBefore": now.Add(-24 * time.Hour).Unix(),
	}, &ids)
	return ids, err
}
