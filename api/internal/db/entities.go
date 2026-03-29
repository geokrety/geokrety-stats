package db

import (
	"strings"
	"time"

	geokrety "github.com/geokrety/geokrety-stats/geokrety/geokrety"
	movetypes "github.com/geokrety/geokrety-stats/geokrety/move"
)

type GeokretListItem struct {
	ID               int64               `db:"id" json:"id" xml:"id"`
	GKID             *geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	Name             string              `db:"name" json:"name" xml:"name"`
	AvatarID         *int64              `db:"avatar_id" json:"avatarId" xml:"avatarId,omitempty"`
	AvatarURL        *string             `db:"avatar_url" json:"avatarUrl,omitempty" xml:"avatarUrl,omitempty"`
	Type             int16               `db:"type" json:"type" xml:"type"`
	TypeName         string              `json:"typeName" xml:"typeName"`
	Missing          bool                `db:"missing" json:"missing" xml:"missing"`
	MissingAt        *time.Time          `db:"missing_at" json:"missingAt" xml:"missingAt,omitempty"`
	OwnerID          *int64              `db:"owner_id" json:"ownerId" xml:"ownerId,omitempty"`
	OwnerUsername    *string             `db:"owner_username" json:"ownerUsername" xml:"ownerUsername,omitempty"`
	HolderID         *int64              `db:"holder_id" json:"holderId" xml:"holderId,omitempty"`
	HolderUsername   *string             `db:"holder_username" json:"holderUsername" xml:"holderUsername,omitempty"`
	Country          *string             `db:"country" json:"country" xml:"country,omitempty"`
	Waypoint         *string             `db:"waypoint" json:"waypoint" xml:"waypoint,omitempty"`
	Lat              *float64            `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon              *float64            `db:"lon" json:"lon" xml:"lon,omitempty"`
	BornAt           *time.Time          `db:"born_at" json:"bornAt" xml:"bornAt,omitempty"`
	LastMoveAt       *time.Time          `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt,omitempty"`
	LastMoveType     *int16              `db:"last_move_type" json:"lastMoveType,omitempty" xml:"lastMoveType,omitempty"`
	LastPositionID   *int64              `db:"last_position_id" json:"lastPositionId,omitempty" xml:"lastPositionId,omitempty"`
	LastLogID        *int64              `db:"last_log_id" json:"lastLogId,omitempty" xml:"lastLogId,omitempty"`
	Mission          *string             `db:"mission" json:"mission" xml:"mission,omitempty"`
	NonCollectibleAt *time.Time          `db:"non_collectible_at" json:"nonCollectibleAt,omitempty" xml:"nonCollectibleAt,omitempty"`
	ParkedAt         *time.Time          `db:"parked_at" json:"parkedAt,omitempty" xml:"parkedAt,omitempty"`
	CommentsHidden   bool                `db:"comments_hidden" json:"commentsHidden" xml:"commentsHidden"`
	GeoJSON          *GeoJSONPt          `json:"geojson" xml:"geojson,omitempty"`
}

type GeokretDetails struct {
	GeokretListItem
}

type GeokretStats struct {
	GeokretID             int64               `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	GKID                  *geokrety.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	CachesCount           int64               `db:"caches_count" json:"cachesCount" xml:"cachesCount"`
	PicturesCount         int64               `db:"pictures_count" json:"picturesCount" xml:"picturesCount"`
	LovesCount            int64               `db:"loves_count" json:"lovesCount" xml:"lovesCount"`
	MovesCount            int64               `db:"moves_count" json:"movesCount" xml:"movesCount"`
	CountriesVisitedCount int64               `db:"countries_visited_count" json:"countriesVisitedCount" xml:"countriesVisitedCount"`
	WaypointsVisitedCount int64               `db:"waypoints_visited_count" json:"waypointsVisitedCount" xml:"waypointsVisitedCount"`
	FindersCount          int64               `db:"finders_count" json:"findersCount" xml:"findersCount"`
	WatchersCount         int64               `db:"watchers_count" json:"watchersCount" xml:"watchersCount"`
	LoversCount           int64               `db:"lovers_count" json:"loversCount" xml:"loversCount"`
	CurrentCountryCode    *string             `db:"current_country_code" json:"currentCountryCode,omitempty" xml:"currentCountryCode,omitempty"`
	CurrentWaypointCode   *string             `db:"current_waypoint_code" json:"currentWaypointCode,omitempty" xml:"currentWaypointCode,omitempty"`
}

type MoveFilters struct {
	GeokretID *int64
	UserID    *int64
	Country   *string
	Waypoint  *string
	DateFrom  *time.Time
	DateTo    *time.Time
}

type MoveRecord struct {
	ID                 int64               `db:"id" json:"id" xml:"id"`
	GeokretID          int64               `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	GeokretGKID        *geokrety.GeokretId `db:"geokret_gkid" json:"geokretGkid,omitempty" xml:"geokretGkid,omitempty"`
	MoveType           int16               `db:"move_type" json:"moveType" xml:"moveType"`
	MoveTypeName       string              `json:"moveTypeName" xml:"moveTypeName"`
	AuthorID           *int64              `db:"author_id" json:"authorId" xml:"authorId,omitempty"`
	AuthorAvatarID     *int64              `db:"author_avatar_id" json:"authorAvatarId,omitempty" xml:"authorAvatarId,omitempty"`
	AuthorAvatarURL    *string             `db:"author_avatar_url" json:"authorAvatarUrl,omitempty" xml:"authorAvatarUrl,omitempty"`
	Username           string              `db:"username" json:"username" xml:"username"`
	Country            *string             `db:"country" json:"country" xml:"country,omitempty"`
	Waypoint           *string             `db:"waypoint" json:"waypoint" xml:"waypoint,omitempty"`
	Lat                *float64            `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon                *float64            `db:"lon" json:"lon" xml:"lon,omitempty"`
	Elevation          *int64              `db:"elevation" json:"elevation" xml:"elevation,omitempty"`
	KMDistance         *float64            `db:"km_distance" json:"kmDistance" xml:"kmDistance,omitempty"`
	MovedOn            time.Time           `db:"moved_on_datetime" json:"movedOn" xml:"movedOn"`
	CreatedOn          time.Time           `db:"created_on_datetime" json:"createdOn" xml:"createdOn"`
	Comment            *string             `db:"comment" json:"comment" xml:"comment,omitempty"`
	CommentHidden      bool                `db:"comment_hidden" json:"commentHidden" xml:"commentHidden"`
	PreviousMoveID     *int64              `db:"previous_move_id" json:"previousMoveId,omitempty" xml:"previousMoveId,omitempty"`
	PreviousPositionID *int64              `db:"previous_position_id" json:"previousPositionId,omitempty" xml:"previousPositionId,omitempty"`
	GeoJSON            *GeoJSONPt          `json:"geojson" xml:"geojson,omitempty"`
}

type SocialUserEntry struct {
	UserID    int64     `db:"user_id" json:"userId" xml:"userId"`
	Username  string    `db:"username" json:"username" xml:"username"`
	AvatarID  *int64    `db:"avatar_id" json:"avatarId,omitempty" xml:"avatarId,omitempty"`
	AvatarURL *string   `db:"avatar_url" json:"avatarUrl,omitempty" xml:"avatarUrl,omitempty"`
	At        time.Time `db:"at" json:"at" xml:"at"`
}

type PictureFilters struct {
	GeokretID *int64
	MoveID    *int64
	UserID    *int64
}

type PictureInfo struct {
	ID             int64               `db:"id" json:"id" xml:"id"`
	Type           int16               `db:"type" json:"type" xml:"type"`
	Filename       *string             `db:"filename" json:"filename" xml:"filename,omitempty"`
	Caption        *string             `db:"caption" json:"caption" xml:"caption,omitempty"`
	Key            *string             `db:"key" json:"key" xml:"key,omitempty"`
	GeokretID      *int64              `db:"geokret_id" json:"geokretId" xml:"geokretId,omitempty"`
	GeokretGKID    *geokrety.GeokretId `db:"geokret_gkid" json:"geokretGkid,omitempty" xml:"geokretGkid,omitempty"`
	MoveID         *int64              `db:"move_id" json:"moveId" xml:"moveId,omitempty"`
	UserID         *int64              `db:"user_id" json:"userId" xml:"userId,omitempty"`
	AuthorID       *int64              `db:"author_id" json:"authorId" xml:"authorId,omitempty"`
	AuthorUsername *string             `db:"author_username" json:"authorUsername" xml:"authorUsername,omitempty"`
	UploadedOn     *time.Time          `db:"uploaded_on_datetime" json:"uploadedOn" xml:"uploadedOn,omitempty"`
	CreatedOn      time.Time           `db:"created_on_datetime" json:"createdOn" xml:"createdOn"`
}

type GeokretCountryVisit struct {
	CountryCode    string    `db:"country_code" json:"countryCode" xml:"countryCode"`
	FirstVisitedAt time.Time `db:"first_visited_at" json:"firstVisitedAt" xml:"firstVisitedAt"`
	MoveCount      int64     `db:"move_count" json:"moveCount" xml:"moveCount"`
	Flag           string    `json:"flag" xml:"flag"`
}

type GeokretWaypointVisit struct {
	WaypointCode   string     `db:"waypoint_code" json:"waypointCode" xml:"waypointCode"`
	Country        *string    `db:"country" json:"country" xml:"country,omitempty"`
	VisitCount     int64      `db:"visit_count" json:"visitCount" xml:"visitCount"`
	FirstVisitedAt time.Time  `db:"first_visited_at" json:"firstVisitedAt" xml:"firstVisitedAt"`
	LastVisitedAt  time.Time  `db:"last_visited_at" json:"lastVisitedAt" xml:"lastVisitedAt"`
	Lat            *float64   `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon            *float64   `db:"lon" json:"lon" xml:"lon,omitempty"`
	GeoJSON        *GeoJSONPt `json:"geojson" xml:"geojson,omitempty"`
}

type CountryDetails struct {
	Code          string  `db:"code" json:"code" xml:"code"`
	Name          string  `db:"name" json:"name" xml:"name"`
	ContinentCode *string `db:"continent_code" json:"continentCode" xml:"continentCode,omitempty"`
	ContinentName *string `db:"continent_name" json:"continentName" xml:"continentName,omitempty"`
	Flag          string  `json:"flag" xml:"flag"`
}

type WaypointSummary struct {
	ID           int64      `db:"id" json:"id" xml:"id"`
	WaypointCode string     `db:"waypoint_code" json:"waypointCode" xml:"waypointCode"`
	Source       string     `db:"source" json:"source" xml:"source"`
	Country      *string    `db:"country" json:"country" xml:"country,omitempty"`
	Lat          *float64   `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon          *float64   `db:"lon" json:"lon" xml:"lon,omitempty"`
	GeoJSON      *GeoJSONPt `json:"geojson" xml:"geojson,omitempty"`
}

type WaypointDetails struct {
	WaypointSummary
}

type UserDetails struct {
	ID              int64      `db:"id" json:"id" xml:"id"`
	Username        string     `db:"username" json:"username" xml:"username"`
	JoinedAt        time.Time  `db:"joined_at" json:"joinedAt" xml:"joinedAt"`
	HomeCountry     *string    `db:"home_country" json:"homeCountry" xml:"homeCountry,omitempty"`
	AvatarID        *int64     `db:"avatar_id" json:"avatarId" xml:"avatarId,omitempty"`
	AvatarURL       *string    `db:"avatar_url" json:"avatarUrl,omitempty" xml:"avatarUrl,omitempty"`
	LastMoveAt      *time.Time `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt,omitempty"`
	HomeCountryFlag string     `json:"homeCountryFlag" xml:"homeCountryFlag"`
}

type UserStats struct {
	UserID                int64 `db:"user_id" json:"userId" xml:"userId"`
	OwnedGeokretyCount    int64 `db:"owned_geokrety_count" json:"ownedGeokretyCount" xml:"ownedGeokretyCount"`
	FoundGeokretyCount    int64 `db:"found_geokrety_count" json:"foundGeokretyCount" xml:"foundGeokretyCount"`
	LovedGeokretyCount    int64 `db:"loved_geokrety_count" json:"lovedGeokretyCount" xml:"lovedGeokretyCount"`
	WatchedGeokretyCount  int64 `db:"watched_geokrety_count" json:"watchedGeokretyCount" xml:"watchedGeokretyCount"`
	PicturesCount         int64 `db:"pictures_count" json:"picturesCount" xml:"picturesCount"`
	CountriesVisitedCount int64 `db:"countries_visited_count" json:"countriesVisitedCount" xml:"countriesVisitedCount"`
	WaypointsVisitedCount int64 `db:"waypoints_visited_count" json:"waypointsVisitedCount" xml:"waypointsVisitedCount"`
	MovesCount            int64 `db:"moves_count" json:"movesCount" xml:"movesCount"`
	DistinctGeokretyCount int64 `db:"distinct_geokrety_count" json:"distinctGeokretyCount" xml:"distinctGeokretyCount"`
}

type UserCountryVisit struct {
	CountryCode string    `db:"country_code" json:"countryCode" xml:"countryCode"`
	MoveCount   int64     `db:"move_count" json:"moveCount" xml:"moveCount"`
	FirstVisit  time.Time `db:"first_visit" json:"firstVisit" xml:"firstVisit"`
	LastVisit   time.Time `db:"last_visit" json:"lastVisit" xml:"lastVisit"`
	Flag        string    `json:"flag" xml:"flag"`
}

type UserWaypointVisit struct {
	WaypointCode   string     `db:"waypoint_code" json:"waypointCode" xml:"waypointCode"`
	VisitCount     int64      `db:"visit_count" json:"visitCount" xml:"visitCount"`
	FirstVisitedAt time.Time  `db:"first_visited_at" json:"firstVisitedAt" xml:"firstVisitedAt"`
	LastVisitedAt  time.Time  `db:"last_visited_at" json:"lastVisitedAt" xml:"lastVisitedAt"`
	Country        *string    `db:"country" json:"country" xml:"country,omitempty"`
	Lat            *float64   `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon            *float64   `db:"lon" json:"lon" xml:"lon,omitempty"`
	GeoJSON        *GeoJSONPt `json:"geojson" xml:"geojson,omitempty"`
}

type UserSearchResult struct {
	ID              int64      `db:"id" json:"id" xml:"id"`
	Username        string     `db:"username" json:"username" xml:"username"`
	JoinedAt        time.Time  `db:"joined_at" json:"joinedAt" xml:"joinedAt"`
	HomeCountry     *string    `db:"home_country" json:"homeCountry" xml:"homeCountry,omitempty"`
	AvatarID        *int64     `db:"avatar_id" json:"avatarId,omitempty" xml:"avatarId,omitempty"`
	AvatarURL       *string    `db:"avatar_url" json:"avatarUrl,omitempty" xml:"avatarUrl,omitempty"`
	LastMoveAt      *time.Time `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt,omitempty"`
	HomeCountryFlag string     `json:"homeCountryFlag" xml:"homeCountryFlag"`
}

type GeoJSONPt struct {
	Type        string    `json:"type" xml:"type"`
	Coordinates []float64 `json:"coordinates" xml:"coordinates>coordinate"`
}

func hydrateGeokretListItems(items []GeokretListItem) []GeokretListItem {
	for i := range items {
		items[i].TypeName = geokrety.DefaultGeokretTypeRegistry.Name(items[i].Type)
		if !items[i].Missing {
			items[i].MissingAt = nil
		}
		if items[i].Country != nil {
			country := strings.ToUpper(*items[i].Country)
			items[i].Country = &country
		}
		if items[i].Waypoint != nil {
			waypoint := strings.ToUpper(*items[i].Waypoint)
			items[i].Waypoint = &waypoint
		}
		if items[i].Lat != nil && items[i].Lon != nil {
			items[i].GeoJSON = &GeoJSONPt{Type: "Point", Coordinates: []float64{*items[i].Lon, *items[i].Lat}}
		}
	}
	return items
}

func hydrateMoveRecords(items []MoveRecord) []MoveRecord {
	for i := range items {
		items[i].MoveTypeName = movetypes.DefaultMoveTypeRegistry.Name(items[i].MoveType)
		if items[i].Country != nil {
			country := strings.ToUpper(*items[i].Country)
			items[i].Country = &country
		}
		if items[i].Waypoint != nil {
			waypoint := strings.ToUpper(*items[i].Waypoint)
			items[i].Waypoint = &waypoint
		}
		if items[i].Lat != nil && items[i].Lon != nil {
			items[i].GeoJSON = &GeoJSONPt{Type: "Point", Coordinates: []float64{*items[i].Lon, *items[i].Lat}}
		}
	}
	return items
}

func hydrateWaypointSummary(item WaypointSummary) WaypointSummary {
	if item.Country != nil {
		country := strings.ToUpper(*item.Country)
		item.Country = &country
	}
	if item.Lat != nil && item.Lon != nil {
		item.GeoJSON = &GeoJSONPt{Type: "Point", Coordinates: []float64{*item.Lon, *item.Lat}}
	}
	return item
}

func hydrateWaypointVisits(items []UserWaypointVisit) []UserWaypointVisit {
	for i := range items {
		if items[i].Country != nil {
			country := strings.ToUpper(*items[i].Country)
			items[i].Country = &country
		}
		if items[i].Lat != nil && items[i].Lon != nil {
			items[i].GeoJSON = &GeoJSONPt{Type: "Point", Coordinates: []float64{*items[i].Lon, *items[i].Lat}}
		}
	}
	return items
}

func hydrateGKWaypointVisits(items []GeokretWaypointVisit) []GeokretWaypointVisit {
	for i := range items {
		if items[i].Country != nil {
			country := strings.ToUpper(*items[i].Country)
			items[i].Country = &country
		}
		if items[i].Lat != nil && items[i].Lon != nil {
			items[i].GeoJSON = &GeoJSONPt{Type: "Point", Coordinates: []float64{*items[i].Lon, *items[i].Lat}}
		}
	}
	return items
}

func hydrateUserRows(items []UserSearchResult) []UserSearchResult {
	for i := range items {
		if items[i].HomeCountry != nil {
			country := strings.ToUpper(*items[i].HomeCountry)
			items[i].HomeCountry = &country
			items[i].HomeCountryFlag = countryFlag(country)
		}
	}
	return items
	
}

