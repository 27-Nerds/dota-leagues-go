package model

// Update is a public activity-feed entry recorded after a successful data write.
type Update struct {
	ID             string         `json:"id"`
	Entity         string         `json:"entity"`
	EntityID       int            `json:"entity_id"`
	Name           string         `json:"name"`
	Action         string         `json:"action"`
	CreatedAt      int64          `json:"created_at"`
	URL            string         `json:"url,omitempty"`
	Team           *UpdateTeam    `json:"team,omitempty"`
	Changes        []UpdateChange `json:"changes"`
	RosterAdmins   map[int]bool   `json:"roster_admins,omitempty"`
	GroupID        string         `json:"group_id,omitempty"`
	GroupKind      string         `json:"group_kind,omitempty"`
	GroupStartedAt int64          `json:"group_started_at,omitempty"`
	Sources        []Update       `json:"sources,omitempty"`
}

// UpdateChange contains the previous and current value of a public field.
type UpdateChange struct {
	Field  string `json:"field"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

// UpdateTeam preserves the player’s team when the event was recorded.
type UpdateTeam struct {
	Available bool   `json:"available"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
}

type UpdateFilter struct {
	Search string
	Entity string
	Days   int
}
