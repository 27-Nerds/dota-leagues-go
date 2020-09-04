package model

// TeamRoster model
type TeamRoster struct {
	TeamID      int          `json:"team_id"`
	TeamMembers []TeamMember `json:"team_members"`
	DbKey       string       `json:"_key,omitempty"`
}

// TeamMember model part
type TeamMember struct {
	AccountID int  `json:"account_id"`
	IsActive  bool `json:"is_active"`
}
