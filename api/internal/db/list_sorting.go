package db

func safeOrderBy(sort Sort, allowed map[string]string, fallback Sort) string {
	normalized := sort.String()
	if clause, ok := allowed[normalized]; ok {
		return clause
	}
	return allowed[fallback.String()]
}

func geokretOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":              "last_move_at DESC NULLS LAST, id DESC",
		"last_move_at":  "last_move_at ASC NULLS LAST, id ASC",
		"-last_move_at": "last_move_at DESC NULLS LAST, id DESC",
		"name":          "name ASC, id ASC",
		"-name":         "name DESC, id DESC",
		"born_at":       "born_at ASC NULLS LAST, id ASC",
		"-born_at":      "born_at DESC NULLS LAST, id DESC",
	}, DescSort("last_move_at"))
}

func moveOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":      "moved_on_datetime DESC, id DESC",
		"date":  "moved_on_datetime ASC, id ASC",
		"-date": "moved_on_datetime DESC, id DESC",
		"id":    "id ASC",
		"-id":   "id DESC",
	}, DescSort("date"))
}

func socialOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":                 "at DESC, user_id DESC",
		"at":               "at ASC, user_id ASC",
		"-at":              "at DESC, user_id DESC",
		"loved_on_date":    "at ASC, user_id ASC",
		"-loved_on_date":   "at DESC, user_id DESC",
		"watched_on_date":  "at ASC, user_id ASC",
		"-watched_on_date": "at DESC, user_id DESC",
		"found_on_date":    "at ASC, user_id ASC",
		"-found_on_date":   "at DESC, user_id DESC",
	}, DescSort("at"))
}

func countryOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":      "code ASC",
		"code":  "code ASC",
		"-code": "code DESC",
		"name":  "name ASC, code ASC",
		"-name": "name DESC, code DESC",
	}, AscSort("code"))
}

func userOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":              "joined_at DESC, id DESC",
		"username":      "username ASC, id ASC",
		"-username":     "username DESC, id DESC",
		"joined_at":     "joined_at ASC, id ASC",
		"-joined_at":    "joined_at DESC, id DESC",
		"last_move_at":  "last_move_at ASC NULLS LAST, id ASC",
		"-last_move_at": "last_move_at DESC NULLS LAST, id DESC",
	}, DescSort("joined_at"))
}

func pictureOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":            "created_on_datetime DESC, id DESC",
		"created_on":  "created_on_datetime ASC, id ASC",
		"-created_on": "created_on_datetime DESC, id DESC",
		"id":          "id ASC",
		"-id":         "id DESC",
	}, DescSort("created_on"))
}

func geokretCountryVisitOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":                  "first_visited_at DESC, country_code ASC",
		"country_code":      "country_code ASC",
		"-country_code":     "country_code DESC",
		"first_visited_at":  "first_visited_at ASC, country_code ASC",
		"-first_visited_at": "first_visited_at DESC, country_code ASC",
		"move_count":        "move_count ASC, country_code ASC",
		"-move_count":       "move_count DESC, country_code ASC",
	}, DescSort("first_visited_at"))
}

func userCountryVisitOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":              "last_visit DESC, country_code ASC",
		"country_code":  "country_code ASC",
		"-country_code": "country_code DESC",
		"first_visit":   "first_visit ASC, country_code ASC",
		"-first_visit":  "first_visit DESC, country_code ASC",
		"last_visit":    "last_visit ASC, country_code ASC",
		"-last_visit":   "last_visit DESC, country_code ASC",
		"move_count":    "move_count ASC, country_code ASC",
		"-move_count":   "move_count DESC, country_code ASC",
	}, DescSort("last_visit"))
}

func waypointVisitOrderBy(sort Sort) string {
	return safeOrderBy(sort, map[string]string{
		"":                  "last_visited_at DESC, waypoint_code ASC",
		"waypoint_code":     "waypoint_code ASC",
		"-waypoint_code":    "waypoint_code DESC",
		"first_visited_at":  "first_visited_at ASC, waypoint_code ASC",
		"-first_visited_at": "first_visited_at DESC, waypoint_code ASC",
		"last_visited_at":   "last_visited_at ASC, waypoint_code ASC",
		"-last_visited_at":  "last_visited_at DESC, waypoint_code ASC",
		"visit_count":       "visit_count ASC, waypoint_code ASC",
		"-visit_count":      "visit_count DESC, waypoint_code ASC",
	}, DescSort("last_visited_at"))
}
