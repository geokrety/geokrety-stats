package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/geokrety/geokrety-stats-api/internal/db"
	sharedjsonrest "github.com/geokrety/geokrety-stats/geokrety/jsonrest"
)

type resourceDescriptor struct {
	ResourceType string
	IDFields     []string
	CompositeID  bool
}

var resourceDescriptors = map[reflect.Type]resourceDescriptor{
	reflect.TypeFor[db.GlobalStats]():                 {ResourceType: "global_stats", IDFields: []string{"global"}},
	reflect.TypeFor[db.CountryStats]():                {ResourceType: "country_stat", IDFields: []string{"Code"}},
	reflect.TypeFor[db.LeaderboardUser]():             {ResourceType: "leaderboard_entry", IDFields: []string{"UserID"}},
	reflect.TypeFor[db.RecentMove]():                  {ResourceType: "move", IDFields: []string{"ID"}},
	reflect.TypeFor[db.RecentBorn]():                  {ResourceType: "geokret", IDFields: []string{"GKID", "ID"}},
	reflect.TypeFor[db.RecentLoved]():                 {ResourceType: "love", IDFields: []string{"GKID", "GeoKretID", "UserID", "LovedAt"}},
	reflect.TypeFor[db.RecentWatched]():               {ResourceType: "watch", IDFields: []string{"GKID", "GeoKretID", "UserID", "WatchedAt"}},
	reflect.TypeFor[db.ActiveCountry]():               {ResourceType: "country_activity", IDFields: []string{"Code"}},
	reflect.TypeFor[db.ActiveWaypoint]():              {ResourceType: "waypoint_activity", IDFields: []string{"Waypoint"}},
	reflect.TypeFor[db.RecentRegisteredUser]():        {ResourceType: "user", IDFields: []string{"ID"}},
	reflect.TypeFor[db.RecentActiveUser]():            {ResourceType: "user_activity", IDFields: []string{"UserID"}},
	reflect.TypeFor[db.HourlyHeatmapCell]():           {ResourceType: "hourly_heatmap_cell", IDFields: []string{"ActivityDate", "HourUTC", "MoveType"}, CompositeID: true},
	reflect.TypeFor[db.CountryFlow]():                 {ResourceType: "country_flow", IDFields: []string{"YearMonth", "FromCountry", "ToCountry"}, CompositeID: true},
	reflect.TypeFor[db.TopCache]():                    {ResourceType: "top_cache", IDFields: []string{"WaypointCode"}},
	reflect.TypeFor[db.FirstFinderLeaderboardEntry](): {ResourceType: "first_finder_entry", IDFields: []string{"UserID"}},
	reflect.TypeFor[db.DistanceRecord]():              {ResourceType: "distance_record", IDFields: []string{"GKID", "GeoKretID"}},
	reflect.TypeFor[db.UserNetworkEdge]():             {ResourceType: "user_network_edge", IDFields: []string{"UserID", "RelatedUserID"}, CompositeID: true},
	reflect.TypeFor[db.GeokretTimelineEvent]():        {ResourceType: "geokret_event", IDFields: []string{"GeoKretID", "EventType", "OccurredAt"}, CompositeID: true},
	reflect.TypeFor[db.GeokretCirculation]():          {ResourceType: "geokret_circulation", IDFields: []string{"GKID", "GeoKretID"}},
	reflect.TypeFor[db.GeokretListItem]():             {ResourceType: "geokret", IDFields: []string{"GKID", "ID"}},
	reflect.TypeFor[db.GeokretDetails]():              {ResourceType: "geokret", IDFields: []string{"GKID", "ID"}},
	reflect.TypeFor[db.MoveRecord]():                  {ResourceType: "move", IDFields: []string{"ID"}},
	reflect.TypeFor[db.SocialUserEntry]():             {ResourceType: "user", IDFields: []string{"UserID"}},
	reflect.TypeFor[db.PictureInfo]():                 {ResourceType: "picture", IDFields: []string{"ID"}},
	reflect.TypeFor[db.GeokretCountryVisit]():         {ResourceType: "country_visit", IDFields: []string{"CountryCode"}},
	reflect.TypeFor[db.GeokretWaypointVisit]():        {ResourceType: "waypoint_visit", IDFields: []string{"WaypointCode"}},
	reflect.TypeFor[db.CountryCount]():                {ResourceType: "country_stat", IDFields: []string{"CountryCode"}},
	reflect.TypeFor[db.ElevationPoint]():              {ResourceType: "elevation_point", IDFields: []string{"MoveID"}},
	reflect.TypeFor[db.DayHeatmapCell]():              {ResourceType: "day_heatmap_cell", IDFields: []string{"Day"}},
	reflect.TypeFor[db.HourHeatmapCell]():             {ResourceType: "hour_heatmap_cell", IDFields: []string{"DayOfWeek", "HourUTC"}, CompositeID: true},
	reflect.TypeFor[db.DormancyRecord]():              {ResourceType: "dormancy_record", IDFields: []string{"GKID", "GeokretID"}},
	reflect.TypeFor[db.MultiplierVelocityRecord]():    {ResourceType: "multiplier_velocity_record", IDFields: []string{"GKID", "GeokretID"}},
	reflect.TypeFor[db.UserContinentCoverage]():       {ResourceType: "continent_coverage", IDFields: []string{"UserID", "ContinentCode"}, CompositeID: true},
	reflect.TypeFor[db.TripPoint]():                   {ResourceType: "trip_point", IDFields: []string{"MoveID"}},
	reflect.TypeFor[db.CountryDetails]():              {ResourceType: "country", IDFields: []string{"Code"}},
	reflect.TypeFor[db.WaypointSummary]():             {ResourceType: "waypoint", IDFields: []string{"WaypointCode", "ID"}},
	reflect.TypeFor[db.WaypointDetails]():             {ResourceType: "waypoint", IDFields: []string{"WaypointCode", "ID"}},
	reflect.TypeFor[db.UserDetails]():                 {ResourceType: "user", IDFields: []string{"ID"}},
	reflect.TypeFor[db.UserCountryVisit]():            {ResourceType: "country_visit", IDFields: []string{"CountryCode"}},
	reflect.TypeFor[db.UserWaypointVisit]():           {ResourceType: "waypoint_visit", IDFields: []string{"WaypointCode"}},
	reflect.TypeFor[db.UserSearchResult]():            {ResourceType: "user", IDFields: []string{"ID"}},
}

func resourceDataFromPayload(payload any) any {
	if payload == nil {
		return nil
	}
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
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		resources := make([]sharedjsonrest.Resource, 0, value.Len())
		for index := range value.Len() {
			resource, ok := resourceFromValue(value.Index(index))
			if !ok {
				return payload
			}
			resources = append(resources, resource)
		}
		return resources
	case reflect.Struct:
		resource, ok := resourceFromValue(value)
		if !ok {
			return payload
		}
		return resource
	default:
		return payload
	}
}

func resourceFromValue(value reflect.Value) (sharedjsonrest.Resource, bool) {
	value = dereferenceValue(value)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return sharedjsonrest.Resource{}, false
	}
	descriptor, ok := resourceDescriptors[value.Type()]
	if !ok {
		descriptor = resourceDescriptor{ResourceType: toSnakeCase(value.Type().Name()), IDFields: []string{"ID", "UserID", "GKID", "Code", "CountryCode", "WaypointCode", "MoveID"}}
	}
	id := buildResourceID(value, descriptor)
	if id == "" {
		return sharedjsonrest.Resource{}, false
	}
	attributes := make(map[string]any)
	appendStructAttributes(attributes, value, toSkipSet(descriptor.IDFields))
	return sharedjsonrest.Resource{
		ID:         id,
		Type:       descriptor.ResourceType,
		Attributes: attributes,
	}, true
}

func appendStructAttributes(target map[string]any, value reflect.Value, skip map[string]struct{}) {
	for _, field := range reflect.VisibleFields(value.Type()) {
		if field.PkgPath != "" {
			continue
		}
		fieldValue := value.FieldByIndex(field.Index)
		if field.Anonymous {
			embedded := dereferenceValue(fieldValue)
			if embedded.IsValid() && embedded.Kind() == reflect.Struct {
				appendStructAttributes(target, embedded, skip)
			}
			continue
		}
		if _, shouldSkip := skip[field.Name]; shouldSkip {
			continue
		}
		encoded, ok := encodeAttributeValue(fieldValue)
		if !ok {
			continue
		}
		target[attributeName(field)] = encoded
	}
}

func encodeAttributeValue(value reflect.Value) (any, bool) {
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

func attributeName(field reflect.StructField) string {
	if dbTag := strings.TrimSpace(strings.Split(field.Tag.Get("db"), ",")[0]); dbTag != "" && dbTag != "-" {
		return dbTag
	}
	return toSnakeCase(field.Name)
}

func buildResourceID(value reflect.Value, descriptor resourceDescriptor) string {
	if len(descriptor.IDFields) == 1 && descriptor.IDFields[0] == "global" {
		return "global"
	}
	parts := make([]string, 0, len(descriptor.IDFields))
	for _, name := range descriptor.IDFields {
		part := fieldString(value, name)
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return ""
	}
	if !descriptor.CompositeID {
		return parts[0]
	}
	return strings.Join(parts, ":")
}

func fieldString(value reflect.Value, fieldName string) string {
	fieldValue, ok := findField(value, fieldName)
	if !ok {
		return ""
	}
	fieldValue = dereferenceValue(fieldValue)
	if !fieldValue.IsValid() {
		return ""
	}
	if fieldValue.Type() == reflect.TypeFor[time.Time]() {
		return fieldValue.Interface().(time.Time).UTC().Format(time.RFC3339)
	}
	encoded, ok := encodeAttributeValue(fieldValue)
	if !ok || encoded == nil {
		return ""
	}
	return fmt.Sprint(encoded)
}

func findField(value reflect.Value, fieldName string) (reflect.Value, bool) {
	for _, field := range reflect.VisibleFields(value.Type()) {
		if field.Name != fieldName {
			continue
		}
		return value.FieldByIndex(field.Index), true
	}
	return reflect.Value{}, false
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

func toSkipSet(fields []string) map[string]struct{} {
	skip := make(map[string]struct{}, len(fields)+1)
	for _, field := range fields {
		skip[field] = struct{}{}
	}
	if _, hasGKID := skip["GKID"]; hasGKID {
		skip["ID"] = struct{}{}
	}
	if _, hasGeoKretID := skip["GeoKretID"]; hasGeoKretID {
		skip["GKID"] = struct{}{}
	}
	return skip
}

func toSnakeCase(value string) string {
	if value == "" {
		return ""
	}
	var builder strings.Builder
	for index, char := range value {
		if unicode.IsUpper(char) {
			if index > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(unicode.ToLower(char))
			continue
		}
		builder.WriteRune(char)
	}
	return builder.String()
}
