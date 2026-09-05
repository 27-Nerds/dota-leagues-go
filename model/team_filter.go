package model

// TeamFilter selects directory entries and their ordering.
type TeamFilter struct {
	Search     string
	Country    string
	Pro        *bool
	Active     bool
	ActiveDays int
	Sort       string
	Order      string
}
