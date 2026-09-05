package model

import "encoding/json"

type SourceSnapshot struct {
	Key       string          `json:"_key"`
	Kind      string          `json:"kind"`
	EntityID  int             `json:"entity_id"`
	Hash      string          `json:"hash"`
	FirstSeen int64           `json:"first_seen"`
	LastSeen  int64           `json:"last_seen"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}
type SourceStatus struct {
	Key         string `json:"_key"`
	Kind        string `json:"kind"`
	EntityID    int    `json:"entity_id"`
	Name        string `json:"name"`
	LastAttempt int64  `json:"last_attempt"`
	LastSuccess int64  `json:"last_success"`
	Failure     string `json:"failure"`
	Versions    int    `json:"versions"`
	AuditCount  int    `json:"audit_count"`
	GroupCount  int    `json:"group_count"`
}
type SourceInspection struct {
	PlayerNames map[string]string `json:"player_names"`
	Status      *SourceStatus     `json:"status"`
	Versions    []SourceSnapshot  `json:"versions"`
	Snapshot    *SourceSnapshot   `json:"snapshot"`
}
