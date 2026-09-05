package repository

// Sort fragments are selected from fixed expressions, never interpolated input.
func sortOrder(sort, order string) string {
	if order == "asc" || (order == "" && sort == "name") {
		return "ASC"
	}
	return "DESC"
}
