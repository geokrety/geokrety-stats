package handlers

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	sharedjsonrest "github.com/geokrety/geokrety-stats/geokrety/jsonrest"
)

func resourceDataFromPayload(r *http.Request, payload any) any {
	switch typed := dereferencePayload(payload).(type) {
	case db.GeokretDetails:
		return presentGeokretResource(typed.GeokretListItem)
	case []db.GeokretListItem:
		return presentGeokretCollection(typed)
	case db.GeokretStats:
		return presentGeokretStatsResource(typed)
	case []db.MoveRecord:
		return presentMoveCollection(typed)
	case db.MoveRecord:
		return presentMoveResource(typed)
	case []db.SocialUserEntry:
		return presentSocialUserCollection(typed)
	case []db.PictureInfo:
		return presentPictureCollection(typed)
	case db.PictureInfo:
		return presentPictureResource(typed)
	case []db.GeokretCountryVisit:
		return presentGeokretCountryVisitCollection(typed)
	case []db.GeokretWaypointVisit:
		return presentGeokretWaypointVisitCollection(typed)
	case []db.CountryDetails:
		return presentCountryCollection(typed)
	case db.CountryDetails:
		return presentCountryResource(typed)
	case []db.WaypointSummary:
		return presentWaypointSummaryCollection(typed)
	case db.WaypointDetails:
		return presentWaypointDetailsResource(typed)
	case []db.UserSearchResult:
		return presentUserSearchCollection(typed)
	case db.UserDetails:
		return presentUserDetailsResource(typed)
	case db.UserStats:
		return presentUserStatsResource(typed)
	case []db.UserCountryVisit:
		return presentUserCountryVisitCollection(typed)
	case []db.UserWaypointVisit:
		return presentUserWaypointVisitCollection(typed)
	default:
		return payload
	}
}

func presentGeokretCollection(rows []db.GeokretListItem) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentGeokretResource(row))
	}
	return resources
}

func presentGeokretResource(row db.GeokretListItem) sharedjsonrest.Resource {
	gkid := gkidString(row.GKID)
	selfPath := geokretPath(gkid)
	attributes := map[string]any{
		"name":            row.Name,
		"type_id":         row.Type,
		"type_name":       row.TypeName,
		"missing":         row.Missing,
		"collectible":     row.NonCollectibleAt == nil,
		"parked":          row.ParkedAt != nil,
		"comments_hidden": row.CommentsHidden,
	}
	setInt64Ptr(attributes, "avatar_id", row.AvatarID)
	setStringPtr(attributes, "avatar_url", row.AvatarURL)
	setTimePtr(attributes, "missing_at", row.MissingAt)
	setTimePtr(attributes, "born_at", row.BornAt)
	setTimePtr(attributes, "last_move_at", row.LastMoveAt)
	setStringPtr(attributes, "mission", row.Mission)
	if row.GeoJSON != nil {
		attributes["position"] = row.GeoJSON
	}

	relationships := relationships(
		relationshipWithResource("owner", embeddedUserResource(row.OwnerID, row.OwnerUsername, nil, nil), userPath(int64PtrString(row.OwnerID))),
		relationshipWithResource("holder", embeddedUserResource(row.HolderID, row.HolderUsername, nil, nil), userPath(int64PtrString(row.HolderID))),
		relationshipWithResource("country", embeddedCountryResource(row.Country, nil), countryPath(stringPtrString(row.Country))),
		relationshipWithResource("waypoint", embeddedWaypointResource(row.Waypoint), waypointPath(stringPtrString(row.Waypoint))),
		relationshipWithIdentifier("last_position", "move", int64PtrString(row.LastPositionID), movePath(int64PtrString(row.LastPositionID))),
		relationshipWithIdentifier("last_log", "move", int64PtrString(row.LastLogID), movePath(int64PtrString(row.LastLogID))),
		collectionRelationship("moves", selfPath+"/moves"),
		collectionRelationship("avatars", selfPath+"/pictures"),
		collectionRelationship("lovers", selfPath+"/loved-by"),
		collectionRelationship("watchers", selfPath+"/watched-by"),
		collectionRelationship("finders", selfPath+"/finders"),
		collectionRelationship("countries_visited", selfPath+"/countries"),
		collectionRelationship("waypoints_visited", selfPath+"/waypoints"),
		collectionRelationship("stats", selfPath+"/stats"),
	)

	return newResource(gkid, "geokrety", mapOrNil(attributes), relationships, linksForPath(selfPath))
}

func presentGeokretStatsResource(row db.GeokretStats) sharedjsonrest.Resource {
	gkid := gkidString(row.GKID)
	selfPath := geokretPath(gkid) + "/stats"
	attributes := map[string]any{
		"caches_count":            row.CachesCount,
		"pictures_count":          row.PicturesCount,
		"loves_count":             row.LovesCount,
		"moves_count":             row.MovesCount,
		"countries_visited_count": row.CountriesVisitedCount,
		"waypoints_visited_count": row.WaypointsVisitedCount,
		"finders_count":           row.FindersCount,
		"watchers_count":          row.WatchersCount,
		"lovers_count":            row.LoversCount,
	}
	relationships := relationships(
		relationshipWithIdentifier("geokret", "geokrety", gkid, geokretPath(gkid)),
		relationshipWithResource("current_country", embeddedCountryResource(row.CurrentCountryCode, nil), countryPath(stringPtrString(row.CurrentCountryCode))),
		relationshipWithResource("current_waypoint", embeddedWaypointResource(row.CurrentWaypointCode), waypointPath(stringPtrString(row.CurrentWaypointCode))),
	)
	return newResource(gkid, "geokrety_stats", attributes, relationships, linksForPath(selfPath))
}

func presentMoveCollection(rows []db.MoveRecord) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentMoveResource(row))
	}
	return resources
}

func presentMoveResource(row db.MoveRecord) sharedjsonrest.Resource {
	moveID := strconv.FormatInt(row.ID, 10)
	selfPath := movePath(moveID)
	attributes := map[string]any{
		"type_id":        row.MoveType,
		"type":           moveTypeSlug(row.MoveType),
		"date":           row.MovedOn.UTC().Format(time.RFC3339),
		"username":       nonEmptyString(row.Username, "unknown"),
		"comment_hidden": row.CommentHidden,
		"created_on":     row.CreatedOn.UTC().Format(time.RFC3339),
	}
	setStringPtr(attributes, "comment", row.Comment)
	setInt64Ptr(attributes, "elevation", row.Elevation)
	setFloat64Ptr(attributes, "km_distance", row.KMDistance)
	if row.GeoJSON != nil {
		attributes["position"] = row.GeoJSON
	}

	relationships := relationships(
		relationshipWithResource("geokret", embeddedGeokretResource(row.GeokretGKID, nil), geokretPath(gkidString(row.GeokretGKID))),
		relationshipWithResource("author", embeddedUserResource(row.AuthorID, localStringPtr(nonEmptyString(row.Username, "")), row.AuthorAvatarID, row.AuthorAvatarURL), userPath(int64PtrString(row.AuthorID))),
		relationshipWithResource("country", embeddedCountryResource(row.Country, nil), countryPath(stringPtrString(row.Country))),
		relationshipWithResource("waypoint", embeddedWaypointResource(row.Waypoint), waypointPath(stringPtrString(row.Waypoint))),
		relationshipWithIdentifier("previous_move", "move", int64PtrString(row.PreviousMoveID), movePath(int64PtrString(row.PreviousMoveID))),
		relationshipWithIdentifier("previous_position", "move", int64PtrString(row.PreviousPositionID), movePath(int64PtrString(row.PreviousPositionID))),
		collectionRelationship("pictures", "/api/v3/pictures?move="+moveID),
	)
	return newResource(moveID, "move", mapOrNil(attributes), relationships, linksForPath(selfPath))
}

func presentSocialUserCollection(rows []db.SocialUserEntry) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		userID := strconv.FormatInt(row.UserID, 10)
		attributes := map[string]any{
			"username": row.Username,
			"at":       row.At.UTC().Format(time.RFC3339),
		}
		setInt64Ptr(attributes, "avatar_id", row.AvatarID)
		setStringPtr(attributes, "avatar_url", row.AvatarURL)
		resources = append(resources, newResource(userID, "user", mapOrNil(attributes), nil, linksForPath(userPath(userID))))
	}
	return resources
}

func presentPictureCollection(rows []db.PictureInfo) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentPictureResource(row))
	}
	return resources
}

func presentPictureResource(row db.PictureInfo) sharedjsonrest.Resource {
	pictureID := strconv.FormatInt(row.ID, 10)
	attributes := map[string]any{
		"type":       row.Type,
		"created_on": row.CreatedOn.UTC().Format(time.RFC3339),
	}
	setStringPtr(attributes, "filename", row.Filename)
	setStringPtr(attributes, "caption", row.Caption)
	setStringPtr(attributes, "key", row.Key)
	setStringPtr(attributes, "author_username", row.AuthorUsername)
	setTimePtr(attributes, "uploaded_on", row.UploadedOn)

	relationships := relationships(
		relationshipWithResource("geokret", embeddedGeokretResource(row.GeokretGKID, nil), geokretPath(gkidString(row.GeokretGKID))),
		relationshipWithIdentifier("move", "move", int64PtrString(row.MoveID), movePath(int64PtrString(row.MoveID))),
		relationshipWithIdentifier("user", "user", int64PtrString(row.UserID), userPath(int64PtrString(row.UserID))),
		relationshipWithIdentifier("author", "user", int64PtrString(row.AuthorID), userPath(int64PtrString(row.AuthorID))),
	)
	return newResource(pictureID, "picture", mapOrNil(attributes), relationships, linksForPath("/api/v3/pictures/"+pictureID))
}

func presentGeokretCountryVisitCollection(rows []db.GeokretCountryVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		attributes := map[string]any{
			"first_visited_at": row.FirstVisitedAt.UTC().Format(time.RFC3339),
			"move_count":       row.MoveCount,
			"flag":             row.Flag,
		}
		resources = append(resources, newResource(
			row.CountryCode,
			"country_visit",
			attributes,
			relationships(relationshipWithResource("country", embeddedCountryResource(&row.CountryCode, localStringPtr(row.Flag)), countryPath(row.CountryCode))),
			linksForPath(countryPath(row.CountryCode)),
		))
	}
	return resources
}

func presentGeokretWaypointVisitCollection(rows []db.GeokretWaypointVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		attributes := map[string]any{
			"visit_count":      row.VisitCount,
			"first_visited_at": row.FirstVisitedAt.UTC().Format(time.RFC3339),
			"last_visited_at":  row.LastVisitedAt.UTC().Format(time.RFC3339),
		}
		if row.GeoJSON != nil {
			attributes["position"] = row.GeoJSON
		}
		resources = append(resources, newResource(
			row.WaypointCode,
			"waypoint_visit",
			mapOrNil(attributes),
			relationships(
				relationshipWithResource("waypoint", embeddedWaypointResource(&row.WaypointCode), waypointPath(row.WaypointCode)),
				relationshipWithResource("country", embeddedCountryResource(row.Country, nil), countryPath(stringPtrString(row.Country))),
			),
			linksForPath(waypointPath(row.WaypointCode)),
		))
	}
	return resources
}

func presentCountryCollection(rows []db.CountryDetails) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentCountryResource(row))
	}
	return resources
}

func presentCountryResource(row db.CountryDetails) sharedjsonrest.Resource {
	attributes := map[string]any{
		"name": row.Name,
		"flag": row.Flag,
	}
	setStringPtr(attributes, "continent_code", row.ContinentCode)
	setStringPtr(attributes, "continent_name", row.ContinentName)
	relationships := relationships(
		collectionRelationship("geokrety", "/api/v3/countries/"+row.Code+"/geokrety"),
	)
	return newResource(row.Code, "country", mapOrNil(attributes), relationships, linksForPath(countryPath(row.Code)))
}

func presentWaypointSummaryCollection(rows []db.WaypointSummary) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentWaypointSummaryResource(row))
	}
	return resources
}

func presentWaypointDetailsResource(row db.WaypointDetails) sharedjsonrest.Resource {
	return presentWaypointSummaryResource(row.WaypointSummary)
}

func presentWaypointSummaryResource(row db.WaypointSummary) sharedjsonrest.Resource {
	attributes := map[string]any{
		"source": row.Source,
	}
	if row.GeoJSON != nil {
		attributes["position"] = row.GeoJSON
	}
	relationships := relationships(
		relationshipWithResource("country", embeddedCountryResource(row.Country, nil), countryPath(stringPtrString(row.Country))),
		collectionRelationship("current_geokrety", "/api/v3/waypoints/"+row.WaypointCode+"/geokrety-current"),
		collectionRelationship("past_geokrety", "/api/v3/waypoints/"+row.WaypointCode+"/geokrety-past"),
	)
	return newResource(row.WaypointCode, "waypoint", mapOrNil(attributes), relationships, linksForPath(waypointPath(row.WaypointCode)))
}

func presentUserSearchCollection(rows []db.UserSearchResult) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentUserResource(row.ID, row.Username, row.JoinedAt, row.LastMoveAt, row.AvatarID, row.AvatarURL, row.HomeCountry, row.HomeCountryFlag))
	}
	return resources
}

func presentUserDetailsResource(row db.UserDetails) sharedjsonrest.Resource {
	return presentUserResource(row.ID, row.Username, row.JoinedAt, row.LastMoveAt, row.AvatarID, row.AvatarURL, row.HomeCountry, row.HomeCountryFlag)
}

func presentUserResource(id int64, username string, joinedAt time.Time, lastMoveAt *time.Time, avatarID *int64, avatarURL *string, homeCountry *string, homeCountryFlag string) sharedjsonrest.Resource {
	userID := strconv.FormatInt(id, 10)
	attributes := map[string]any{
		"username":  username,
		"joined_at": joinedAt.UTC().Format("2006-01-02"),
	}
	setInt64Ptr(attributes, "avatar_id", avatarID)
	setStringPtr(attributes, "avatar_url", avatarURL)
	setDatePtr(attributes, "last_move_at", lastMoveAt)
	relationships := relationships(
		relationshipWithResource("home_country", embeddedCountryResource(homeCountry, localStringPtr(homeCountryFlag)), countryPath(stringPtrString(homeCountry))),
		collectionRelationship("owned_geokrety", userPath(userID)+"/geokrety-owned"),
		collectionRelationship("found_geokrety", userPath(userID)+"/geokrety-found"),
		collectionRelationship("loved_geokrety", userPath(userID)+"/geokrety-loved"),
		collectionRelationship("watched_geokrety", userPath(userID)+"/geokrety-watched"),
		collectionRelationship("pictures", userPath(userID)+"/pictures"),
		collectionRelationship("countries_visited", userPath(userID)+"/countries"),
		collectionRelationship("waypoints_visited", userPath(userID)+"/waypoints"),
		collectionRelationship("stats", userPath(userID)+"/stats"),
	)
	return newResource(userID, "user", mapOrNil(attributes), relationships, linksForPath(userPath(userID)))
}

func presentUserStatsResource(row db.UserStats) sharedjsonrest.Resource {
	userID := strconv.FormatInt(row.UserID, 10)
	attributes := map[string]any{
		"owned_geokrety_count":    row.OwnedGeokretyCount,
		"found_geokrety_count":    row.FoundGeokretyCount,
		"loved_geokrety_count":    row.LovedGeokretyCount,
		"watched_geokrety_count":  row.WatchedGeokretyCount,
		"pictures_count":          row.PicturesCount,
		"countries_visited_count": row.CountriesVisitedCount,
		"waypoints_visited_count": row.WaypointsVisitedCount,
		"moves_count":             row.MovesCount,
		"distinct_geokrety_count": row.DistinctGeokretyCount,
	}
	relationships := relationships(
		relationshipWithIdentifier("user", "user", userID, userPath(userID)),
	)
	return newResource(userID, "user_stats", attributes, relationships, linksForPath(userPath(userID)+"/stats"))
}

func presentUserCountryVisitCollection(rows []db.UserCountryVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		attributes := map[string]any{
			"move_count":  row.MoveCount,
			"first_visit": row.FirstVisit.UTC().Format(time.RFC3339),
			"last_visit":  row.LastVisit.UTC().Format(time.RFC3339),
			"flag":        row.Flag,
		}
		resources = append(resources, newResource(
			row.CountryCode,
			"country_visit",
			attributes,
			relationships(relationshipWithResource("country", embeddedCountryResource(&row.CountryCode, localStringPtr(row.Flag)), countryPath(row.CountryCode))),
			linksForPath(countryPath(row.CountryCode)),
		))
	}
	return resources
}

func presentUserWaypointVisitCollection(rows []db.UserWaypointVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		attributes := map[string]any{
			"visit_count":      row.VisitCount,
			"first_visited_at": row.FirstVisitedAt.UTC().Format(time.RFC3339),
			"last_visited_at":  row.LastVisitedAt.UTC().Format(time.RFC3339),
		}
		if row.GeoJSON != nil {
			attributes["position"] = row.GeoJSON
		}
		resources = append(resources, newResource(
			row.WaypointCode,
			"waypoint_visit",
			mapOrNil(attributes),
			relationships(
				relationshipWithResource("waypoint", embeddedWaypointResource(&row.WaypointCode), waypointPath(row.WaypointCode)),
				relationshipWithResource("country", embeddedCountryResource(row.Country, nil), countryPath(stringPtrString(row.Country))),
			),
			linksForPath(waypointPath(row.WaypointCode)),
		))
	}
	return resources
}

func newResource(id, resourceType string, attributes map[string]any, relationships map[string]sharedjsonrest.Relationship, links sharedjsonrest.Links) sharedjsonrest.Resource {
	return sharedjsonrest.Resource{
		ID:            id,
		Type:          resourceType,
		Attributes:    attributes,
		Relationships: relationships,
		Links:         links,
	}
}

type namedRelationship struct {
	Name         string
	Relationship sharedjsonrest.Relationship
}

func relationships(values ...namedRelationship) map[string]sharedjsonrest.Relationship {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]sharedjsonrest.Relationship)
	for _, value := range values {
		if value.Name == "" {
			continue
		}
		if value.Relationship.Data == nil && len(value.Relationship.Links) == 0 {
			continue
		}
		result[value.Name] = value.Relationship
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func collectionRelationship(name, relatedPath string) namedRelationship {
	if relatedPath == "" {
		return namedRelationship{}
	}
	return namedRelationship{
		Name: name,
		Relationship: sharedjsonrest.Relationship{
			Links: relatedLinks(relatedPath),
		},
	}
}

func relationshipWithIdentifier(name, resourceType, id, relatedPath string) namedRelationship {
	if id == "" && relatedPath == "" {
		return namedRelationship{}
	}
	relationship := sharedjsonrest.Relationship{Links: relatedLinks(relatedPath)}
	if id != "" {
		relationship.Data = sharedjsonrest.Identifier{Type: resourceType, ID: id}
	}
	return namedRelationship{Name: name, Relationship: relationship}
}

func relationshipWithResource(name string, resource *sharedjsonrest.Resource, relatedPath string) namedRelationship {
	if resource == nil && relatedPath == "" {
		return namedRelationship{}
	}
	relationship := sharedjsonrest.Relationship{Links: relatedLinks(relatedPath)}
	if resource != nil {
		relationship.Data = *resource
	}
	return namedRelationship{Name: name, Relationship: relationship}
}

func embeddedUserResource(id *int64, username *string, avatarID *int64, avatarURL *string) *sharedjsonrest.Resource {
	if id == nil {
		return nil
	}
	userID := strconv.FormatInt(*id, 10)
	attributes := map[string]any{}
	setStringPtr(attributes, "username", username)
	setInt64Ptr(attributes, "avatar_id", avatarID)
	setStringPtr(attributes, "avatar_url", avatarURL)
	resource := newResource(userID, "user", mapOrNil(attributes), nil, linksForPath(userPath(userID)))
	return &resource
}

func embeddedCountryResource(code *string, flag *string) *sharedjsonrest.Resource {
	if code == nil || strings.TrimSpace(*code) == "" {
		return nil
	}
	countryCode := strings.ToUpper(strings.TrimSpace(*code))
	attributes := map[string]any{}
	resolvedFlag := countryFlagFromCode(countryCode)
	if flag != nil && strings.TrimSpace(*flag) != "" {
		resolvedFlag = *flag
	}
	setString(attributes, "flag", resolvedFlag)
	resource := newResource(countryCode, "country", mapOrNil(attributes), nil, linksForPath(countryPath(countryCode)))
	return &resource
}

func embeddedWaypointResource(code *string) *sharedjsonrest.Resource {
	if code == nil || strings.TrimSpace(*code) == "" {
		return nil
	}
	waypointCode := strings.ToUpper(strings.TrimSpace(*code))
	resource := newResource(waypointCode, "waypoint", nil, nil, linksForPath(waypointPath(waypointCode)))
	return &resource
}

func embeddedGeokretResource(gkid *geokrety.GeokretId, name *string) *sharedjsonrest.Resource {
	id := gkidString(gkid)
	if id == "" {
		return nil
	}
	attributes := map[string]any{}
	setStringPtr(attributes, "name", name)
	resource := newResource(id, "geokrety", mapOrNil(attributes), nil, linksForPath(geokretPath(id)))
	return &resource
}

func dereferencePayload(payload any) any {
	value := reflect.ValueOf(payload)
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	return value.Interface()
}

func linksForPath(path string) sharedjsonrest.Links {
	if path == "" {
		return nil
	}
	return sharedjsonrest.Links{}.Set("self", path)
}

func relatedLinks(path string) sharedjsonrest.Links {
	if path == "" {
		return nil
	}
	return sharedjsonrest.Links{}.Set("related", path)
}

func geokretPath(gkid string) string {
	if gkid == "" {
		return ""
	}
	return "/api/v3/geokrety/" + gkid
}

func movePath(moveID string) string {
	if moveID == "" {
		return ""
	}
	return "/api/v3/moves/" + moveID
}

func userPath(userID string) string {
	if userID == "" {
		return ""
	}
	return "/api/v3/users/" + userID
}

func countryPath(code string) string {
	if strings.TrimSpace(code) == "" {
		return ""
	}
	return "/api/v3/countries/" + strings.ToUpper(strings.TrimSpace(code))
}

func waypointPath(code string) string {
	if strings.TrimSpace(code) == "" {
		return ""
	}
	return "/api/v3/waypoints/" + strings.ToUpper(strings.TrimSpace(code))
}

func gkidString(gkid *geokrety.GeokretId) string {
	if gkid == nil {
		return ""
	}
	return gkid.ToGKIDOrEmpty()
}

func int64PtrString(value *int64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(*value, 10)
}

func stringPtrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func localStringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	copyValue := value
	return &copyValue
}

func mapOrNil(attributes map[string]any) map[string]any {
	if len(attributes) == 0 {
		return nil
	}
	return attributes
}

func setString(target map[string]any, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	target[key] = value
}

func setStringPtr(target map[string]any, key string, value *string) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return
	}
	target[key] = *value
}

func setInt64Ptr(target map[string]any, key string, value *int64) {
	if value == nil {
		return
	}
	target[key] = *value
}

func setFloat64Ptr(target map[string]any, key string, value *float64) {
	if value == nil {
		return
	}
	target[key] = *value
}

func setTimePtr(target map[string]any, key string, value *time.Time) {
	if value == nil {
		return
	}
	target[key] = value.UTC().Format(time.RFC3339)
}

func setDatePtr(target map[string]any, key string, value *time.Time) {
	if value == nil {
		return
	}
	target[key] = value.UTC().Format("2006-01-02")
}

func countryFlagFromCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 2 {
		return ""
	}
	runes := []rune(code)
	if runes[0] < 'A' || runes[0] > 'Z' || runes[1] < 'A' || runes[1] > 'Z' {
		return ""
	}
	return string([]rune{runes[0] - 'A' + 0x1F1E6, runes[1] - 'A' + 0x1F1E6})
}

func moveTypeSlug(moveType int16) string {
	switch moveType {
	case 0:
		return "drop"
	case 1:
		return "grab"
	case 2:
		return "comment"
	case 3:
		return "seen"
	case 4:
		return "archived"
	case 5:
		return "dip"
	default:
		return "unknown"
	}
}

func nonEmptyString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
