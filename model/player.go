package model

// PlayersData - player api response
type PlayersData struct {
	Players []Player `json:"player_infos"`
}

// PlayerResult is one tournament placement from the DPC profile, enriched with the league name on read.
type PlayerResult struct {
	LeagueID        int    `json:"league_id"`
	Placement       int    `json:"placement"`
	Earnings        int    `json:"earnings"`
	LeagueName      string `json:"league_name,omitempty"`
	LeagueAvailable bool   `json:"league_available,omitempty"`
}

// ProRegistration is one entry of Valve's professional registration history.
type ProRegistration struct {
	RegistrationPeriod int `json:"registration_period"`
	Timestamp          int `json:"timestamp"`
}

// PlayerTeamEntry is one team membership from Valve's per-player audit log.
type PlayerTeamEntry struct {
	StartTimestamp int    `json:"start_timestamp"`
	TeamID         int    `json:"team_id"`
	TeamName       string `json:"team_name,omitempty"`
	TeamTag        string `json:"team_tag,omitempty"`
	TeamURLLogo    string `json:"team_url_logo,omitempty"`
	TeamAvailable  bool   `json:"team_available,omitempty"`
}

// PlayerRosterTeam is the stored roster listing with the latest joining date.
// It is independent of the DPC profile and does not establish an active team.
type PlayerRosterTeam struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Tag         string `json:"tag,omitempty"`
	JoinedAt    int    `json:"joined_at,omitempty"`
	RetrievedAt int64  `json:"retrieved_at,omitempty"`
}

// Player - player api and db struct
type Player struct {
	SteamName          string `json:"steam_name,omitempty"`
	SteamUpdatedAt     int64  `json:"steam_updated_at,omitempty"`
	SteamNextRefreshAt int64  `json:"steam_next_refresh_at,omitempty"`
	SteamStatus        string `json:"steam_status,omitempty"`
	// SteamPrivacy is the profile visibility Steam reported on the last successful lookup.
	SteamPrivacy string `json:"steam_privacy,omitempty"`
	// SteamCheckedAt records the last lookup attempt, successful or not.
	SteamCheckedAt int64 `json:"steam_checked_at,omitempty"`
	// SteamPublicAt is the last lookup that found the profile public; zero when never seen public.
	SteamPublicAt int64 `json:"steam_public_at,omitempty"`
	// SteamLocation is the self-reported profile location, not nationality.
	SteamLocation    string            `json:"steam_location,omitempty"`
	ProfileSource    string            `json:"profile_source,omitempty"`
	ProfileCheckedAt int64             `json:"profile_checked_at,omitempty"`
	AvatarURL        string            `json:"avatar_url,omitempty"`
	ID               int               `json:"account_id"`
	Name             string            `json:"name"`
	CountryCode      string            `json:"country_code"`
	FantasyRole      int               `json:"fantasy_role"`
	TeamID           int               `json:"team_id"`
	TeamName         string            `json:"team_name,omitempty"`
	TeamTag          string            `json:"team_tag,omitempty"`
	TeamAvailable    bool              `json:"team_available,omitempty"`
	RosterTeam       *PlayerRosterTeam `json:"roster_team,omitempty"`
	// DPCSeenAt is the last sync in which Valve's pro player feed listed this account.
	DPCSeenAt int64 `json:"dpc_seen_at,omitempty"`
	// ProRegistration and TeamHistory come from the pro player feed; IsPro is no longer sent by Valve.
	ProRegistration []ProRegistration `json:"pro_registration,omitempty"`
	TeamHistory     []PlayerTeamEntry `json:"audit_entries,omitempty"`
	Sponsor         string            `json:"sponsor"`
	IsLocked        bool              `json:"is_locked"`
	IsPro           bool              `json:"is_pro"`
	RealName        string            `json:"real_name,omitempty"`
	Birthdate       int               `json:"birthdate,omitempty"`
	TotalEarnings   int               `json:"total_earnings"`
	Results         []PlayerResult    `json:"results,omitempty"`
	TeamURLLogo     string            `json:"team_url_logo,omitempty"`
	DBKey           string            `json:"_key,omitempty"`
}
