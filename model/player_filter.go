package model

// PlayerFilter selects and orders the player directory.
type PlayerFilter struct {
	Search  string
	Country string
	Pro     *bool
	HasTeam *bool
	Sort    string
	Order   string
}
