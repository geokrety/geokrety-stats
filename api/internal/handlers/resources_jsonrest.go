package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	sharedjsonrest "github.com/geokrety/geokrety-stats/geokrety/jsonrest"
	"github.com/go-chi/chi/v5"
)

func resourceDataFromPayload(r *http.Request, payload any) any {
	normalized := dereferencePayload(payload)
	switch typed := normalized.(type) {
	case db.GlobalStats:
		return presentGlobalStats(typed)
	case []db.CountryStats:
		return presentCountryStatCollection(typed)
	case []db.LeaderboardUser:
		return presentLeaderboardCollection(typed)
	case []db.RecentMove:
		return presentRecentMoveCollection(r, typed)
	case []db.RecentBorn:
		return presentRecentBornCollection(typed)
	case []db.RecentLoved:
		return presentRecentLovedCollection(typed)
	case []db.RecentWatched:
		return presentRecentWatchedCollection(typed)
	case []db.ActiveCountry:
		return presentActiveCountryCollection(typed)
	case []db.ActiveWaypoint:
		return presentActiveWaypointCollection(typed)
	case []db.RecentRegisteredUser:
		return presentRecentRegisteredUserCollection(typed)
	case []db.RecentActiveUser:
		return presentRecentActiveUserCollection(typed)
	case db.GeokretDetails:
		return presentGeokretDetails(typed)
	case []db.GeokretListItem:
		return presentGeokretListCollection(typed)
	case []db.MoveRecord:
		return presentMoveCollection(r, typed)
	case db.MoveRecord:
		return presentMoveResource(r, typed)
	case []db.SocialUserEntry:
		return presentSocialUserCollection(typed)
	case []db.PictureInfo:
		return presentPictureCollection(typed)
	case db.PictureInfo:
		return presentPictureResource(typed)
	case []db.DistanceRecord:
		return presentDistanceRecordCollection(typed)
	case []db.DormancyRecord:
		return presentDormancyRecordCollection(typed)
	case []db.MultiplierVelocityRecord:
		return presentMultiplierVelocityCollection(typed)
	case db.GeokretCirculation:
		return presentGeokretCirculation(typed)
	case []db.GeokretCountryVisit:
		return presentGeokretCountryVisitCollection(typed)
	case []db.GeokretWaypointVisit:
		return presentGeokretWaypointVisitCollection(typed)
	case db.CountryDetails:
		return presentCountryResource(typed)
	case []db.CountryDetails:
		return presentCountryCollection(typed)
	case db.WaypointDetails:
		return presentWaypointDetails(typed)
	case []db.WaypointSummary:
		return presentWaypointSummaryCollection(typed)
	case []db.UserSearchResult:
		return presentUserSearchCollection(typed)
	case db.UserDetails:
		return presentUserDetails(typed)
	case []db.UserCountryVisit:
		return presentUserCountryVisitCollection(typed)
	case []db.UserWaypointVisit:
		return presentUserWaypointVisitCollection(typed)
	default:
		return payload
	}
}

func presentGlobalStats(stats db.GlobalStats) sharedjsonrest.Resource {
	return newResource("global", "global_stats", attributesFromDBStruct(stats), nil, linksForPath("/api/v3/stats/kpis"))
}

func presentCountryStatCollection(rows []db.CountryStats) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.Code, "country_stat", attributesFromDBStruct(row, "code"), nil, nil))
	}
	return resources
}

func presentLeaderboardCollection(rows []db.LeaderboardUser) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		userID := strconv.FormatInt(row.UserID, 10)
		resources = append(resources, newResource(userID, "leaderboard_entry", attributesFromDBStruct(row, "user_id"), relationships(
			relationshipWithIdentifier("user", "user", userID),
		), linksForPath("/api/v3/users/"+userID)))
	}
	return resources
}

func presentRecentMoveCollection(r *http.Request, rows []db.RecentMove) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentRecentMoveResource(r, row))
	}
	return resources
}

func presentRecentMoveResource(r *http.Request, row db.RecentMove) sharedjsonrest.Resource {
	moveID := strconv.FormatInt(row.ID, 10)
	return newResource(moveID, "move", attributesFromDBStruct(row, "id", "user_id"), relationships(
		relationshipWithIdentifier("geokret", "geokret", gkidOrEmpty(row.GeokretGKID)),
		relationshipWithIdentifier("user", "user", int64PtrString(row.UserID)),
		relationshipWithIdentifier("country", "country", nonEmptyString(row.Country)),
	), linksForPath(moveSelfPath(r, moveID)))
}

func presentRecentBornCollection(rows []db.RecentBorn) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		gkid := row.GKID.ToGKIDOrEmpty()
		resources = append(resources, newResource(gkid, "geokret", attributesFromDBStruct(row, "id", "gkid", "owner_id"), relationships(
			relationshipWithIdentifier("owner", "user", int64PtrString(row.OwnerID)),
		), linksForPath("/api/v3/geokrety/"+gkid)))
	}
	return resources
}

func presentRecentLovedCollection(rows []db.RecentLoved) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(gkidOrEmpty(row.GKID), "love", attributesFromDBStruct(row, "gkid", "geokret_id", "user_id"), relationships(
			relationshipWithIdentifier("geokret", "geokret", gkidOrEmpty(row.GKID)),
			relationshipWithIdentifier("user", "user", strconv.FormatInt(row.UserID, 10)),
		), nil))
	}
	return resources
}

func presentRecentWatchedCollection(rows []db.RecentWatched) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(gkidOrEmpty(row.GKID), "watch", attributesFromDBStruct(row, "gkid", "geokret_id", "user_id"), relationships(
			relationshipWithIdentifier("geokret", "geokret", gkidOrEmpty(row.GKID)),
			relationshipWithIdentifier("user", "user", strconv.FormatInt(row.UserID, 10)),
		), nil))
	}
	return resources
}

func presentActiveCountryCollection(rows []db.ActiveCountry) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.Code, "country_activity", attributesFromDBStruct(row, "code"), nil, linksForPath("/api/v3/countries/"+row.Code)))
	}
	return resources
}

func presentActiveWaypointCollection(rows []db.ActiveWaypoint) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.Waypoint, "waypoint_activity", attributesFromDBStruct(row, "waypoint"), relationships(
			relationshipWithIdentifier("country", "country", stringPtrString(row.Country)),
		), linksForPath("/api/v3/waypoints/"+row.Waypoint)))
	}
	return resources
}

func presentRecentRegisteredUserCollection(rows []db.RecentRegisteredUser) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		userID := strconv.FormatInt(row.ID, 10)
		resources = append(resources, newResource(userID, "user", attributesFromDBStruct(row, "id"), relationships(
			relationshipWithIdentifier("home_country", "country", stringPtrString(row.HomeCountry)),
		), linksForPath("/api/v3/users/"+userID)))
	}
	return resources
}

func presentRecentActiveUserCollection(rows []db.RecentActiveUser) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		userID := strconv.FormatInt(row.UserID, 10)
		resources = append(resources, newResource(userID, "user_activity", attributesFromDBStruct(row, "user_id"), nil, linksForPath("/api/v3/users/"+userID)))
	}
	return resources
}

func presentGeokretListCollection(rows []db.GeokretListItem) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentGeokretListResource(row))
	}
	return resources
}

func presentGeokretListResource(row db.GeokretListItem) sharedjsonrest.Resource {
	gkid := gkidOrEmpty(row.GKID)
	return newResource(gkid, "geokret", attributesFromDBStruct(row, "id", "gkid", "owner_id", "holder_id"), geokretRelationships(row.OwnerID, row.HolderID, row.Country, row.Waypoint), linksForPath("/api/v3/geokrety/"+gkid))
}

func presentGeokretDetails(row db.GeokretDetails) sharedjsonrest.Resource {
	gkid := gkidOrEmpty(row.GKID)
	return newResource(gkid, "geokret", attributesFromDBStruct(row, "id", "gkid", "owner_id", "holder_id"), geokretRelationships(row.OwnerID, row.HolderID, row.Country, row.Waypoint), linksForPath("/api/v3/geokrety/"+gkid))
}

func presentMoveCollection(r *http.Request, rows []db.MoveRecord) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, presentMoveResource(r, row))
	}
	return resources
}

func presentMoveResource(r *http.Request, row db.MoveRecord) sharedjsonrest.Resource {
	moveID := strconv.FormatInt(row.ID, 10)
	return newResource(moveID, "move", attributesFromDBStruct(row, "id", "geokret_id", "author_id"), relationships(
		relationshipWithIdentifier("author", "user", int64PtrString(row.AuthorID)),
		relationshipWithIdentifier("country", "country", stringPtrString(row.Country)),
		relationshipWithIdentifier("waypoint", "waypoint", stringPtrString(row.Waypoint)),
	), linksForPath(moveSelfPath(r, moveID)))
}

func presentSocialUserCollection(rows []db.SocialUserEntry) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		userID := strconv.FormatInt(row.UserID, 10)
		resources = append(resources, newResource(userID, "user", attributesFromDBStruct(row, "user_id"), nil, linksForPath("/api/v3/users/"+userID)))
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
	return newResource(pictureID, "picture", attributesFromDBStruct(row, "id", "geokret_id", "geokret_gkid", "move_id", "user_id", "author_id"), relationships(
		relationshipWithIdentifier("geokret", "geokret", gkidOrEmpty(row.GeokretGKID)),
		relationshipWithIdentifier("move", "move", int64PtrString(row.MoveID)),
		relationshipWithIdentifier("user", "user", int64PtrString(row.UserID)),
		relationshipWithIdentifier("author", "user", int64PtrString(row.AuthorID)),
	), linksForPath("/api/v3/pictures/"+pictureID))
}

func presentDistanceRecordCollection(rows []db.DistanceRecord) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		gkid := gkidOrEmpty(row.GKID)
		resources = append(resources, newResource(gkid, "distance_record", attributesFromDBStruct(row, "gkid", "gk_id"), relationships(
			relationshipWithIdentifier("geokret", "geokret", gkid),
		), linksForPath("/api/v3/geokrety/"+gkid)))
	}
	return resources
}

func presentDormancyRecordCollection(rows []db.DormancyRecord) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		gkid := gkidOrEmpty(row.GKID)
		resources = append(resources, newResource(gkid, "dormancy_record", attributesFromDBStruct(row, "gkid", "geokret_id"), relationships(
			relationshipWithIdentifier("geokret", "geokret", gkid),
		), linksForPath("/api/v3/geokrety/"+gkid)))
	}
	return resources
}

func presentMultiplierVelocityCollection(rows []db.MultiplierVelocityRecord) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		gkid := gkidOrEmpty(row.GKID)
		resources = append(resources, newResource(gkid, "multiplier_velocity_record", attributesFromDBStruct(row, "gkid", "geokret_id"), relationships(
			relationshipWithIdentifier("geokret", "geokret", gkid),
		), linksForPath("/api/v3/geokrety/"+gkid)))
	}
	return resources
}

func presentGeokretCirculation(row db.GeokretCirculation) sharedjsonrest.Resource {
	gkid := gkidOrEmpty(row.GKID)
	return newResource(gkid, "geokret_circulation", attributesFromDBStruct(row, "gkid", "geokrety_id"), relationships(
		relationshipWithIdentifier("geokret", "geokret", gkid),
	), linksForPath("/api/v3/geokrety/"+gkid))
}

func presentGeokretCountryVisitCollection(rows []db.GeokretCountryVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.CountryCode, "country_visit", attributesFromDBStruct(row, "country_code"), nil, linksForPath("/api/v3/countries/"+row.CountryCode)))
	}
	return resources
}

func presentGeokretWaypointVisitCollection(rows []db.GeokretWaypointVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.WaypointCode, "waypoint_visit", attributesFromDBStruct(row, "waypoint_code"), relationships(
			relationshipWithIdentifier("country", "country", stringPtrString(row.Country)),
		), linksForPath("/api/v3/waypoints/"+row.WaypointCode)))
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
	return newResource(row.Code, "country", attributesFromDBStruct(row, "code", "gkid"), relationships(
		relationshipWithIdentifier("continent", "continent", stringPtrString(row.ContinentCode)),
	), linksForPath("/api/v3/countries/"+row.Code))
}

func presentWaypointSummaryCollection(rows []db.WaypointSummary) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.WaypointCode, "waypoint", attributesFromDBStruct(row, "id", "waypoint_code"), relationships(
			relationshipWithIdentifier("country", "country", stringPtrString(row.Country)),
		), linksForPath("/api/v3/waypoints/"+row.WaypointCode)))
	}
	return resources
}

func presentWaypointDetails(row db.WaypointDetails) sharedjsonrest.Resource {
	return newResource(row.WaypointCode, "waypoint", attributesFromDBStruct(row, "id", "waypoint_code"), relationships(
		relationshipWithIdentifier("country", "country", stringPtrString(row.Country)),
	), linksForPath("/api/v3/waypoints/"+row.WaypointCode))
}

func presentUserSearchCollection(rows []db.UserSearchResult) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		userID := strconv.FormatInt(row.ID, 10)
		resources = append(resources, newResource(userID, "user", attributesFromDBStruct(row, "id"), relationships(
			relationshipWithIdentifier("home_country", "country", stringPtrString(row.HomeCountry)),
		), linksForPath("/api/v3/users/"+userID)))
	}
	return resources
}

func presentUserDetails(row db.UserDetails) sharedjsonrest.Resource {
	userID := strconv.FormatInt(row.ID, 10)
	return newResource(userID, "user", attributesFromDBStruct(row, "id"), relationships(
		relationshipWithIdentifier("home_country", "country", stringPtrString(row.HomeCountry)),
	), linksForPath("/api/v3/users/"+userID))
}

func presentUserCountryVisitCollection(rows []db.UserCountryVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.CountryCode, "country_visit", attributesFromDBStruct(row, "country_code"), nil, linksForPath("/api/v3/countries/"+row.CountryCode)))
	}
	return resources
}

func presentUserWaypointVisitCollection(rows []db.UserWaypointVisit) []sharedjsonrest.Resource {
	resources := make([]sharedjsonrest.Resource, 0, len(rows))
	for _, row := range rows {
		resources = append(resources, newResource(row.WaypointCode, "waypoint_visit", attributesFromDBStruct(row, "waypoint_code"), relationships(
			relationshipWithIdentifier("country", "country", stringPtrString(row.Country)),
		), linksForPath("/api/v3/waypoints/"+row.WaypointCode)))
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
		if value.Name == "" || value.Relationship.Data == nil {
			continue
		}
		result[value.Name] = value.Relationship
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func relationshipWithIdentifier(name, resourceType, id string) namedRelationship {
	if id == "" {
		return namedRelationship{}
	}
	return namedRelationship{
		Name: name,
		Relationship: sharedjsonrest.Relationship{
			Data: sharedjsonrest.Identifier{Type: resourceType, ID: id},
		},
	}
}

func geokretRelationships(ownerID, holderID *int64, country, waypoint *string) map[string]sharedjsonrest.Relationship {
	return relationships(
		relationshipWithIdentifier("owner", "user", int64PtrString(ownerID)),
		relationshipWithIdentifier("holder", "user", int64PtrString(holderID)),
		relationshipWithIdentifier("country", "country", stringPtrString(country)),
		relationshipWithIdentifier("waypoint", "waypoint", stringPtrString(waypoint)),
	)
}

func attributesFromDBStruct(value any, omitKeys ...string) map[string]any {
	reflectValue := dereferenceValue(reflect.ValueOf(value))
	if !reflectValue.IsValid() || reflectValue.Kind() != reflect.Struct {
		return nil
	}
	omitted := make(map[string]struct{}, len(omitKeys))
	for _, key := range omitKeys {
		omitted[key] = struct{}{}
	}
	attributes := map[string]any{}
	appendStructAttributes(attributes, reflectValue, omitted)
	if len(attributes) == 0 {
		return nil
	}
	return attributes
}

func appendStructAttributes(target map[string]any, value reflect.Value, omitted map[string]struct{}) {
	for _, field := range reflect.VisibleFields(value.Type()) {
		if field.PkgPath != "" {
			continue
		}
		fieldValue := value.FieldByIndex(field.Index)
		if field.Anonymous {
			embedded := dereferenceValue(fieldValue)
			if embedded.IsValid() && embedded.Kind() == reflect.Struct {
				appendStructAttributes(target, embedded, omitted)
			}
			continue
		}
		key := dbAttributeKey(field)
		if key == "" {
			continue
		}
		if _, skip := omitted[key]; skip {
			continue
		}
		encoded, ok := encodeValue(fieldValue)
		if !ok {
			continue
		}
		target[key] = encoded
	}
}

func encodeValue(value reflect.Value) (any, bool) {
	value = dereferenceValue(value)
	if !value.IsValid() {
		return nil, false
	}
	encoded, err := json.Marshal(value.Interface())
	if err != nil {
		return nil, false
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, false
	}
	return decoded, true
}

func dbAttributeKey(field reflect.StructField) string {
	if dbTag := strings.TrimSpace(strings.Split(field.Tag.Get("db"), ",")[0]); dbTag != "" && dbTag != "-" {
		return dbTag
	}
	return toSnakeCase(field.Name)
}

func dereferencePayload(payload any) any {
	value := dereferenceValue(reflect.ValueOf(payload))
	if !value.IsValid() {
		return nil
	}
	return value.Interface()
}

func dereferenceValue(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func linksForPath(path string) sharedjsonrest.Links {
	if path == "" {
		return nil
	}
	return sharedjsonrest.Links{"self": path}
}

func moveSelfPath(r *http.Request, moveID string) string {
	if r == nil || r.URL == nil {
		return ""
	}
	geokretPath := canonicalGeokretPathFromRequest(r)
	if geokretPath == "" {
		return ""
	}
	if strings.HasSuffix(r.URL.Path, "/moves") {
		return geokretPath + "/moves/" + moveID
	}
	if strings.Contains(r.URL.Path, "/moves/") {
		return geokretPath + "/moves/" + moveID
	}
	return ""
}

func canonicalGeokretPathFromRequest(r *http.Request) string {
	raw := strings.TrimSpace(chi.URLParam(r, "gkid"))
	if raw == "" {
		raw = strings.TrimSpace(chi.URLParam(r, "id"))
	}
	if raw == "" {
		return ""
	}
	parsed, err := geokrety.New(raw)
	if err != nil {
		return ""
	}
	return "/api/v3/geokrety/" + parsed.ToGKIDOrEmpty()
}

func gkidOrEmpty(gkid *geokrety.GeokretId) string {
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

func nonEmptyString(value string) string {
	if value == "" {
		return ""
	}
	return value
}

func compositeID(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, ":")
}

func toSnakeCase(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	runes := []rune(value)
	for index, char := range runes {
		if unicode.IsUpper(char) {
			if index > 0 {
				previous := runes[index-1]
				nextLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
				if unicode.IsLower(previous) || unicode.IsDigit(previous) || nextLower {
					builder.WriteByte('_')
				}
			}
			builder.WriteRune(unicode.ToLower(char))
			continue
		}
		builder.WriteRune(char)
	}
	return builder.String()
}
