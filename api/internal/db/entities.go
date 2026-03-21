package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/gkid"
)

func geokretTypeName(typeID int16) string {
	switch typeID {
	case 0:
		return "Traditional"
	case 1:
		return "Book/CD/DVD..."
	case 2:
		return "Human/Pet"
	case 3:
		return "Coin"
	case 4:
		return "KretyPost"
	case 5:
		return "Pebble"
	case 6:
		return "Car"
	case 7:
		return "Playing card"
	case 8:
		return "Dog tag/pet"
	case 9:
		return "Jigsaw part"
	case 10:
		return "Hidden GeoKret"
	default:
		return "Unknown"
	}
}

func moveTypeName(typeID int16) string {
	switch typeID {
	case 0:
		return "Dropped"
	case 1:
		return "Grabbed"
	case 2:
		return "Commented"
	case 3:
		return "Seen"
	case 4:
		return "Archived"
	case 5:
		return "Dipped"
	default:
		return "Unknown"
	}
}

type GeokretListItem struct {
	ID             int64           `db:"id" json:"id" xml:"id"`
	GKID           *gkid.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	Name           string          `db:"name" json:"name" xml:"name"`
	Type           int16           `db:"type" json:"type" xml:"type"`
	TypeName       string          `json:"typeName" xml:"typeName"`
	Missing        bool            `db:"missing" json:"missing" xml:"missing"`
	MissingAt      *time.Time      `db:"missing_at" json:"missingAt" xml:"missingAt,omitempty"`
	OwnerID        *int64          `db:"owner_id" json:"ownerId" xml:"ownerId,omitempty"`
	OwnerUsername  *string         `db:"owner_username" json:"ownerUsername" xml:"ownerUsername,omitempty"`
	HolderID       *int64          `db:"holder_id" json:"holderId" xml:"holderId,omitempty"`
	HolderUsername *string         `db:"holder_username" json:"holderUsername" xml:"holderUsername,omitempty"`
	Country        *string         `db:"country" json:"country" xml:"country,omitempty"`
	Waypoint       *string         `db:"waypoint" json:"waypoint" xml:"waypoint,omitempty"`
	Lat            *float64        `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon            *float64        `db:"lon" json:"lon" xml:"lon,omitempty"`
	LovesCount     int64           `db:"loves_count" json:"lovesCount" xml:"lovesCount"`
	PicturesCount  int64           `db:"pictures_count" json:"picturesCount" xml:"picturesCount"`
	BornAt         *time.Time      `db:"born_at" json:"bornAt" xml:"bornAt,omitempty"`
	LastMoveAt     *time.Time      `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt,omitempty"`
	GeoJSON        *GeoJSONPt      `json:"geojson" xml:"geojson,omitempty"`
}

type GeokretDetails struct {
	GeokretListItem
	Mission        *string `db:"mission" json:"mission" xml:"mission,omitempty"`
	DistanceKM     int64   `db:"distance" json:"distanceKm" xml:"distanceKm"`
	CachesCount    int64   `db:"caches_count" json:"cachesCount" xml:"cachesCount"`
	CommentsHidden bool    `db:"comments_hidden" json:"commentsHidden" xml:"commentsHidden"`
}

type MoveRecord struct {
	ID            int64      `db:"id" json:"id" xml:"id"`
	GeokretID     int64      `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	MoveType      int16      `db:"move_type" json:"moveType" xml:"moveType"`
	MoveTypeName  string     `json:"moveTypeName" xml:"moveTypeName"`
	AuthorID      *int64     `db:"author_id" json:"authorId" xml:"authorId,omitempty"`
	Username      *string    `db:"username" json:"username" xml:"username,omitempty"`
	Country       *string    `db:"country" json:"country" xml:"country,omitempty"`
	Waypoint      *string    `db:"waypoint" json:"waypoint" xml:"waypoint,omitempty"`
	Lat           *float64   `db:"lat" json:"lat" xml:"lat,omitempty"`
	Lon           *float64   `db:"lon" json:"lon" xml:"lon,omitempty"`
	Elevation     *int64     `db:"elevation" json:"elevation" xml:"elevation,omitempty"`
	KMDistance    *float64   `db:"km_distance" json:"kmDistance" xml:"kmDistance,omitempty"`
	MovedOn       time.Time  `db:"moved_on_datetime" json:"movedOn" xml:"movedOn"`
	CreatedOn     time.Time  `db:"created_on_datetime" json:"createdOn" xml:"createdOn"`
	PicturesCount int64      `db:"pictures_count" json:"picturesCount" xml:"picturesCount"`
	CommentsCount int64      `db:"comments_count" json:"commentsCount" xml:"commentsCount"`
	Comment       *string    `db:"comment" json:"comment" xml:"comment,omitempty"`
	CommentHidden bool       `db:"comment_hidden" json:"commentHidden" xml:"commentHidden"`
	GeoJSON       *GeoJSONPt `json:"geojson" xml:"geojson,omitempty"`
}

type SocialUserEntry struct {
	UserID   int64     `db:"user_id" json:"userId" xml:"userId"`
	Username string    `db:"username" json:"username" xml:"username"`
	At       time.Time `db:"at" json:"at" xml:"at"`
}

type PictureInfo struct {
	ID             int64      `db:"id" json:"id" xml:"id"`
	Type           int16      `db:"type" json:"type" xml:"type"`
	Filename       *string    `db:"filename" json:"filename" xml:"filename,omitempty"`
	Caption        *string    `db:"caption" json:"caption" xml:"caption,omitempty"`
	Key            *string    `db:"key" json:"key" xml:"key,omitempty"`
	GeokretID      *int64     `db:"geokret_id" json:"geokretId" xml:"geokretId,omitempty"`
	MoveID         *int64     `db:"move_id" json:"moveId" xml:"moveId,omitempty"`
	UserID         *int64     `db:"user_id" json:"userId" xml:"userId,omitempty"`
	AuthorID       *int64     `db:"author_id" json:"authorId" xml:"authorId,omitempty"`
	AuthorUsername *string    `db:"author_username" json:"authorUsername" xml:"authorUsername,omitempty"`
	UploadedOn     *time.Time `db:"uploaded_on_datetime" json:"uploadedOn" xml:"uploadedOn,omitempty"`
	CreatedOn      time.Time  `db:"created_on_datetime" json:"createdOn" xml:"createdOn"`
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

type CountryCount struct {
	CountryCode string     `db:"country_code" json:"countryCode" xml:"countryCode"`
	Value       int64      `db:"value" json:"value" xml:"value"`
	FirstAt     *time.Time `db:"first_at" json:"firstAt" xml:"firstAt,omitempty"`
	Flag        string     `json:"flag" xml:"flag"`
}

type ElevationPoint struct {
	MoveID     int64     `db:"move_id" json:"moveId" xml:"moveId"`
	MovedOn    time.Time `db:"moved_on_datetime" json:"movedOn" xml:"movedOn"`
	Elevation  int64     `db:"elevation" json:"elevation" xml:"elevation"`
	KMDistance *float64  `db:"km_distance" json:"kmDistance" xml:"kmDistance,omitempty"`
	Country    *string   `db:"country" json:"country" xml:"country,omitempty"`
	Waypoint   *string   `db:"waypoint" json:"waypoint" xml:"waypoint,omitempty"`
}

type DayHeatmapCell struct {
	Day       time.Time `db:"day" json:"day" xml:"day"`
	MoveCount int64     `db:"move_count" json:"moveCount" xml:"moveCount"`
}

type HourHeatmapCell struct {
	DayOfWeek int   `db:"day_of_week" json:"dayOfWeek" xml:"dayOfWeek"`
	HourUTC   int   `db:"hour_utc" json:"hourUtc" xml:"hourUtc"`
	MoveCount int64 `db:"move_count" json:"moveCount" xml:"moveCount"`
}

type DormancyRecord struct {
	GeokretID       int64           `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	GKID            *gkid.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	GeokretName     string          `db:"geokret_name" json:"geokretName" xml:"geokretName"`
	LastTouch       *time.Time      `db:"last_touch" json:"lastTouch" xml:"lastTouch,omitempty"`
	DormancySeconds int64           `db:"dormancy_seconds" json:"dormancySeconds" xml:"dormancySeconds"`
	DormancyDays    float64         `db:"dormancy_days" json:"dormancyDays" xml:"dormancyDays"`
}

type MultiplierVelocityRecord struct {
	GeokretID   int64           `db:"geokret_id" json:"geokretId" xml:"geokretId"`
	GKID        *gkid.GeokretId `db:"gkid" json:"gkid" xml:"gkid,omitempty"`
	GeokretName string          `db:"geokret_name" json:"geokretName" xml:"geokretName"`
	LastChange  *time.Time      `db:"last_change" json:"lastChange" xml:"lastChange,omitempty"`
	AvgDelta    float64         `db:"avg_delta" json:"avgDelta" xml:"avgDelta"`
}

type UserContinentCoverage struct {
	UserID        int64   `db:"user_id" json:"userId" xml:"userId"`
	Username      string  `db:"username" json:"username" xml:"username"`
	ContinentCode string  `db:"continent_code" json:"continentCode" xml:"continentCode"`
	ContinentName string  `db:"continent_name" json:"continentName" xml:"continentName"`
	Moves         int64   `db:"moves" json:"moves" xml:"moves"`
	Share         float64 `db:"share" json:"share" xml:"share"`
}

type TripPoint struct {
	MoveID       int64      `db:"move_id" json:"moveId" xml:"moveId"`
	MovedOn      time.Time  `db:"moved_on_datetime" json:"movedOn" xml:"movedOn"`
	MoveType     int16      `db:"move_type" json:"moveType" xml:"moveType"`
	MoveTypeName string     `json:"moveTypeName" xml:"moveTypeName"`
	Country      *string    `db:"country" json:"country" xml:"country,omitempty"`
	Waypoint     *string    `db:"waypoint" json:"waypoint" xml:"waypoint,omitempty"`
	Lat          float64    `db:"lat" json:"lat" xml:"lat"`
	Lon          float64    `db:"lon" json:"lon" xml:"lon"`
	GeoJSON      *GeoJSONPt `json:"geojson" xml:"geojson,omitempty"`
}

type CountryDetails struct {
	Code              string     `db:"code" json:"code" xml:"code"`
	Name              string     `db:"name" json:"name" xml:"name"`
	ContinentCode     *string    `db:"continent_code" json:"continentCode" xml:"continentCode,omitempty"`
	ContinentName     *string    `db:"continent_name" json:"continentName" xml:"continentName,omitempty"`
	MovesCount        int64      `db:"moves_count" json:"movesCount" xml:"movesCount"`
	UniqueUsers       int64      `db:"unique_users" json:"uniqueUsers" xml:"uniqueUsers"`
	UniqueGeokrety    int64      `db:"unique_gks" json:"uniqueGeokrety" xml:"uniqueGeokrety"`
	KMContributed     float64    `db:"km_contributed" json:"kmContributed" xml:"kmContributed"`
	PointsContributed float64    `db:"points_contributed" json:"pointsContributed" xml:"pointsContributed"`
	CurrentGeokrety   int64      `db:"current_geokrety" json:"currentGeokrety" xml:"currentGeokrety"`
	LastStatsDate     *time.Time `db:"last_stats_date" json:"lastStatsDate" xml:"lastStatsDate,omitempty"`
	Flag              string     `json:"flag" xml:"flag"`
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
	CurrentGeokrety int64 `db:"current_geokrety" json:"currentGeokrety" xml:"currentGeokrety"`
	PastGeokrety    int64 `db:"past_geokrety" json:"pastGeokrety" xml:"pastGeokrety"`
	GKVisits        int64 `db:"gk_visits" json:"gkVisits" xml:"gkVisits"`
	UserVisits      int64 `db:"user_visits" json:"userVisits" xml:"userVisits"`
}

type UserDetails struct {
	ID                 int64      `db:"id" json:"id" xml:"id"`
	Username           string     `db:"username" json:"username" xml:"username"`
	JoinedAt           time.Time  `db:"joined_at" json:"joinedAt" xml:"joinedAt"`
	HomeCountry        *string    `db:"home_country" json:"homeCountry" xml:"homeCountry,omitempty"`
	AvatarID           *int64     `db:"avatar_id" json:"avatarId" xml:"avatarId,omitempty"`
	PicturesCount      int64      `db:"pictures_count" json:"picturesCount" xml:"picturesCount"`
	OwnedGeokretyCount int64      `db:"owned_geokrety_count" json:"ownedGeokretyCount" xml:"ownedGeokretyCount"`
	MovesCount         int64      `db:"moves_count" json:"movesCount" xml:"movesCount"`
	DistinctGeokrety   int64      `db:"distinct_geokrety_count" json:"distinctGeokretyCount" xml:"distinctGeokretyCount"`
	ActiveCountries    int64      `db:"active_countries_count" json:"activeCountriesCount" xml:"activeCountriesCount"`
	LastMoveAt         *time.Time `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt,omitempty"`
	HomeCountryFlag    string     `json:"homeCountryFlag" xml:"homeCountryFlag"`
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
	ID          int64      `db:"id" json:"id" xml:"id"`
	Username    string     `db:"username" json:"username" xml:"username"`
	JoinedAt    time.Time  `db:"joined_at" json:"joinedAt" xml:"joinedAt"`
	HomeCountry *string    `db:"home_country" json:"homeCountry" xml:"homeCountry,omitempty"`
	LastMoveAt  *time.Time `db:"last_move_at" json:"lastMoveAt" xml:"lastMoveAt,omitempty"`
}

func hydrateGeokretListItems(items []GeokretListItem) []GeokretListItem {
	for i := range items {
		items[i].TypeName = geokretTypeName(items[i].Type)
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
		items[i].MoveTypeName = moveTypeName(items[i].MoveType)
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

func hydrateTripPoints(items []TripPoint) []TripPoint {
	for i := range items {
		items[i].MoveTypeName = moveTypeName(items[i].MoveType)
		items[i].GeoJSON = &GeoJSONPt{Type: "Point", Coordinates: []float64{items[i].Lon, items[i].Lat}}
		if items[i].Country != nil {
			country := strings.ToUpper(*items[i].Country)
			items[i].Country = &country
		}
		if items[i].Waypoint != nil {
			waypoint := strings.ToUpper(*items[i].Waypoint)
			items[i].Waypoint = &waypoint
		}
	}
	return items
}

func (s *Store) FetchGeokrety(ctx context.Context, geokretID int64) (GeokretDetails, error) {
	return s.fetchGeokretyDetails(ctx, "g.id = $1", geokretID, "query geokret details")
}

func (s *Store) FetchGeokretyByGKID(ctx context.Context, gkid int64) (GeokretDetails, error) {
	return s.fetchGeokretyDetails(ctx, "g.gkid = $1", gkid, "query geokret details by gkid")
}

func (s *Store) FetchGeokretyList(ctx context.Context, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	CASE WHEN gg.missing THEN missing_comment.created_on_datetime END AS missing_at,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
LEFT JOIN LATERAL (
	SELECT mc.created_on_datetime
	FROM geokrety.gk_moves_comments AS mc
	WHERE mc.move = g.last_position
		AND mc.type = 1
	ORDER BY mc.created_on_datetime DESC, mc.id DESC
	LIMIT 1
) AS missing_comment ON TRUE
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret list: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) ResolveGeokretID(ctx context.Context, gkid int64) (int64, error) {
	var geokretID int64
	if err := s.db.GetContext(ctx, &geokretID, `
SELECT id
FROM geokrety.gk_geokrety
WHERE gkid = $1
`, gkid); err != nil {
		return 0, fmt.Errorf("resolve geokret id: %w", err)
	}
	return geokretID, nil
}

func (s *Store) fetchGeokretyDetails(ctx context.Context, predicate string, value int64, label string) (GeokretDetails, error) {
	row := GeokretDetails{}
	query := fmt.Sprintf(`
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	CASE WHEN gg.missing THEN missing_comment.created_on_datetime END AS missing_at,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at,
	gg.mission,
	gg.distance,
	gg.caches_count,
	gg.comments_hidden
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
LEFT JOIN LATERAL (
	SELECT mc.created_on_datetime
	FROM geokrety.gk_moves_comments AS mc
	WHERE mc.move = g.last_position
		AND mc.type = 1
	ORDER BY mc.created_on_datetime DESC, mc.id DESC
	LIMIT 1
) AS missing_comment ON TRUE
WHERE %s
`, predicate)
	if err := s.db.GetContext(ctx, &row, query, value); err != nil {
		return GeokretDetails{}, fmt.Errorf("%s: %w", label, err)
	}
	row.GeokretListItem = hydrateGeokretListItems([]GeokretListItem{row.GeokretListItem})[0]
	return row, nil
}

func (s *Store) FetchGeokretyMoves(ctx context.Context, geokretID int64, limit, offset int) ([]MoveRecord, error) {
	rows := []MoveRecord{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	m.id,
	m.geokret AS geokret_id,
	m.move_type,
	m.author AS author_id,
	COALESCE(u.username, m.username) AS username,
	UPPER(m.country) AS country,
	UPPER(m.waypoint) AS waypoint,
	m.lat,
	m.lon,
	NULLIF(m.elevation, -32768)::bigint AS elevation,
	m.km_distance::double precision AS km_distance,
	m.moved_on_datetime,
	m.created_on_datetime,
	m.pictures_count,
	m.comments_count,
	m.comment,
	m.comment_hidden
FROM geokrety.gk_moves AS m
LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
WHERE m.geokret = $1
ORDER BY m.moved_on_datetime DESC, m.id DESC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret moves: %w", err)
	}
	return hydrateMoveRecords(rows), nil
}

func (s *Store) FetchGeokretyMoveDetails(ctx context.Context, geokretID, moveID int64) (MoveRecord, error) {
	row := MoveRecord{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	m.id,
	m.geokret AS geokret_id,
	m.move_type,
	m.author AS author_id,
	COALESCE(u.username, m.username) AS username,
	UPPER(m.country) AS country,
	UPPER(m.waypoint) AS waypoint,
	m.lat,
	m.lon,
	NULLIF(m.elevation, -32768)::bigint AS elevation,
	m.km_distance::double precision AS km_distance,
	m.moved_on_datetime,
	m.created_on_datetime,
	m.pictures_count,
	m.comments_count,
	m.comment,
	m.comment_hidden
FROM geokrety.gk_moves AS m
LEFT JOIN geokrety.gk_users AS u ON u.id = m.author
WHERE m.geokret = $1 AND m.id = $2
`, geokretID, moveID); err != nil {
		return MoveRecord{}, fmt.Errorf("query geokret move details: %w", err)
	}
	return hydrateMoveRecords([]MoveRecord{row})[0], nil
}

func (s *Store) FetchGeokretyLoves(ctx context.Context, geokretID int64, limit, offset int) ([]SocialUserEntry, error) {
	rows := []SocialUserEntry{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	l.user AS user_id,
	COALESCE(u.username, 'unknown') AS username,
	l.created_on_datetime AS at
FROM geokrety.gk_loves AS l
LEFT JOIN geokrety.gk_users AS u ON u.id = l.user
WHERE l.geokret = $1
ORDER BY l.created_on_datetime DESC, l.id DESC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret loves: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyWatches(ctx context.Context, geokretID int64, limit, offset int) ([]SocialUserEntry, error) {
	rows := []SocialUserEntry{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	w.user AS user_id,
	COALESCE(u.username, 'unknown') AS username,
	w.created_on_datetime AS at
FROM geokrety.gk_watched AS w
LEFT JOIN geokrety.gk_users AS u ON u.id = w.user
WHERE w.geokret = $1
ORDER BY w.created_on_datetime DESC, w.id DESC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret watches: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyPictures(ctx context.Context, geokretID int64, limit, offset int) ([]PictureInfo, error) {
	rows := []PictureInfo{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	p.id,
	p.type,
	p.filename,
	p.caption,
	p.key,
	p.geokret AS geokret_id,
	p.move AS move_id,
	p.user AS user_id,
	p.author AS author_id,
	u.username AS author_username,
	p.uploaded_on_datetime,
	p.created_on_datetime
FROM geokrety.gk_pictures AS p
LEFT JOIN geokrety.gk_users AS u ON u.id = p.author
WHERE p.geokret = $1
ORDER BY p.created_on_datetime DESC, p.id DESC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret pictures: %w", err)
	}
	return rows, nil
}

func (s *Store) SearchGeokrety(ctx context.Context, query string, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	CASE WHEN gg.missing THEN missing_comment.created_on_datetime END AS missing_at,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
LEFT JOIN LATERAL (
	SELECT mc.created_on_datetime
	FROM geokrety.gk_moves_comments AS mc
	WHERE mc.move = g.last_position
		AND mc.type = 1
	ORDER BY mc.created_on_datetime DESC, mc.id DESC
	LIMIT 1
) AS missing_comment ON TRUE
WHERE g.name ILIKE '%' || $1 || '%'
	OR CAST(g.gkid AS text) = $1
	OR CAST(g.id AS text) = $1
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $2 OFFSET $3
`, query, limit, offset); err != nil {
		return nil, fmt.Errorf("search geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchGeokretyCountries(ctx context.Context, geokretID int64, limit, offset int) ([]GeokretCountryVisit, error) {
	rows := []GeokretCountryVisit{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	UPPER(country_code) AS country_code,
	first_visited_at,
	move_count::bigint AS move_count
FROM stats.gk_countries_visited
WHERE geokrety_id = $1
ORDER BY first_visited_at DESC, country_code ASC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret countries: %w", err)
	}
	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].CountryCode)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyWaypoints(ctx context.Context, geokretID int64, limit, offset int) ([]GeokretWaypointVisit, error) {
	rows := []GeokretWaypointVisit{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	w.waypoint_code,
	UPPER(w.country) AS country,
	v.visit_count,
	v.first_visited_at,
	v.last_visited_at,
	w.lat,
	w.lon
FROM stats.gk_cache_visits AS v
INNER JOIN stats.waypoints AS w ON w.id = v.waypoint_id
WHERE v.gk_id = $1
ORDER BY v.last_visited_at DESC, w.waypoint_code ASC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret waypoints: %w", err)
	}
	return hydrateGKWaypointVisits(rows), nil
}

func (s *Store) FetchGeokretyStatsMapCountries(ctx context.Context, geokretID int64, limit, offset int) ([]CountryCount, error) {
	rows := []CountryCount{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	UPPER(country_code) AS country_code,
	move_count::bigint AS value,
	first_visited_at AS first_at
FROM stats.gk_countries_visited
WHERE geokrety_id = $1
ORDER BY move_count DESC, country_code ASC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret map countries: %w", err)
	}
	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].CountryCode)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyStatsElevation(ctx context.Context, geokretID int64, limit, offset int) ([]ElevationPoint, error) {
	rows := []ElevationPoint{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	id AS move_id,
	moved_on_datetime,
	elevation::bigint AS elevation,
	km_distance::double precision AS km_distance,
	UPPER(country) AS country,
	UPPER(waypoint) AS waypoint
FROM geokrety.gk_moves
WHERE geokret = $1
	AND elevation <> -32768
ORDER BY moved_on_datetime DESC, id DESC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret elevation points: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyStatsHeatmapDays(ctx context.Context, geokretID int64, limit, offset int) ([]DayHeatmapCell, error) {
	rows := []DayHeatmapCell{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	DATE(moved_on_datetime AT TIME ZONE 'UTC') AS day,
	COUNT(*)::bigint AS move_count
FROM geokrety.gk_moves
WHERE geokret = $1
GROUP BY DATE(moved_on_datetime AT TIME ZONE 'UTC')
ORDER BY day DESC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret heatmap days: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchStatsDormancy(ctx context.Context, limit, offset int) ([]DormancyRecord, error) {
	rows := []DormancyRecord{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	v.geokrety_id AS geokret_id,
	g.gkid,
	COALESCE(g.name, 'Unknown GeoKret') AS geokret_name,
	v.last_touch,
	COALESCE(EXTRACT(EPOCH FROM v.dormancy_interval), 0)::bigint AS dormancy_seconds,
	COALESCE(ROUND((EXTRACT(EPOCH FROM v.dormancy_interval) / 86400.0)::numeric, 2), 0)::double precision AS dormancy_days
FROM stats.v_uc6_dormancy AS v
LEFT JOIN geokrety.gk_geokrety AS g ON g.id = v.geokrety_id
ORDER BY v.dormancy_interval DESC NULLS LAST, v.geokrety_id ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query dormancy records: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchStatsMultiplierVelocity(ctx context.Context, limit, offset int) ([]MultiplierVelocityRecord, error) {
	rows := []MultiplierVelocityRecord{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	v.gk_id AS geokret_id,
	g.gkid,
	COALESCE(g.name, 'Unknown GeoKret') AS geokret_name,
	v.last_change,
	COALESCE(v.avg_delta, 0)::double precision AS avg_delta
FROM stats.v_uc9_multiplier_velocity AS v
LEFT JOIN geokrety.gk_geokrety AS g ON g.id = v.gk_id
WHERE v.gk_id IS NOT NULL
ORDER BY v.last_change DESC NULLS LAST, ABS(v.avg_delta) DESC NULLS LAST, v.gk_id ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query multiplier velocity records: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchGeokretyTripPoints(ctx context.Context, geokretID int64, limit, offset int) ([]TripPoint, error) {
	rows := []TripPoint{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	id AS move_id,
	moved_on_datetime,
	move_type,
	UPPER(country) AS country,
	UPPER(waypoint) AS waypoint,
	lat,
	lon
FROM geokrety.gk_moves
WHERE geokret = $1
	AND lat IS NOT NULL
	AND lon IS NOT NULL
ORDER BY moved_on_datetime ASC, id ASC
LIMIT $2 OFFSET $3
`, geokretID, limit, offset); err != nil {
		return nil, fmt.Errorf("query geokret trip points: %w", err)
	}
	return hydrateTripPoints(rows), nil
}

func (s *Store) FetchCountryDetails(ctx context.Context, countryCode string) (CountryDetails, error) {
	row := CountryDetails{}
	if err := s.db.GetContext(ctx, &row, `
WITH base AS (
	SELECT
		UPPER($1) AS code,
		MIN(wc.country) AS name,
		MAX(cr.continent_code) AS continent_code,
		MAX(cr.continent_name) AS continent_name,
		COALESCE(SUM(cds.moves_count), 0)::bigint AS moves_count,
		COALESCE(SUM(cds.unique_users), 0)::bigint AS unique_users,
		COALESCE(SUM(cds.unique_gks), 0)::bigint AS unique_gks,
		COALESCE(SUM(cds.km_contributed), 0)::double precision AS km_contributed,
		COALESCE(SUM(cds.points_contributed), 0)::double precision AS points_contributed,
		MAX(cds.stats_date)::timestamp AS last_stats_date
	FROM stats.country_daily_stats AS cds
	LEFT JOIN geokrety.gk_waypoints_country AS wc ON UPPER(wc.original) = UPPER($1)
	LEFT JOIN stats.continent_reference AS cr ON cr.country_alpha2 = UPPER($1)::bpchar
	WHERE UPPER(cds.country_code) = UPPER($1)
)
SELECT
	base.code,
	COALESCE(base.name, base.code) AS name,
	base.continent_code,
	base.continent_name,
	base.moves_count,
	base.unique_users,
	base.unique_gks,
	base.km_contributed,
	base.points_contributed,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_geokrety_with_details AS g
		WHERE UPPER(g.country) = UPPER($1)
	) AS current_geokrety,
	base.last_stats_date
FROM base
`, countryCode); err != nil {
		return CountryDetails{}, fmt.Errorf("query country details: %w", err)
	}
	row.Code = strings.ToUpper(row.Code)
	row.Flag = countryFlag(row.Code)
	return row, nil
}

func (s *Store) FetchCountryGeokrety(ctx context.Context, countryCode string, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
WHERE UPPER(g.country) = UPPER($1)
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $2 OFFSET $3
`, countryCode, limit, offset); err != nil {
		return nil, fmt.Errorf("query country geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchCountryList(ctx context.Context, limit, offset int) ([]CountryDetails, error) {
	rows := []CountryDetails{}
	if err := s.db.SelectContext(ctx, &rows, `
WITH names AS (
	SELECT
		UPPER(original) AS code,
		MIN(country) AS name
	FROM geokrety.gk_waypoints_country
	GROUP BY UPPER(original)
)
SELECT
	UPPER(cds.country_code) AS code,
	COALESCE(MIN(n.name), UPPER(cds.country_code)) AS name,
	MAX(cr.continent_code) AS continent_code,
	MAX(cr.continent_name) AS continent_name,
	COALESCE(SUM(cds.moves_count), 0)::bigint AS moves_count,
	COALESCE(SUM(cds.unique_users), 0)::bigint AS unique_users,
	COALESCE(SUM(cds.unique_gks), 0)::bigint AS unique_gks,
	COALESCE(SUM(cds.km_contributed), 0)::double precision AS km_contributed,
	COALESCE(SUM(cds.points_contributed), 0)::double precision AS points_contributed,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_geokrety_with_details AS g
		WHERE UPPER(g.country) = UPPER(cds.country_code)
	) AS current_geokrety,
	MAX(cds.stats_date)::timestamp AS last_stats_date
FROM stats.country_daily_stats AS cds
LEFT JOIN names AS n ON n.code = UPPER(cds.country_code)
LEFT JOIN stats.continent_reference AS cr ON cr.country_alpha2 = UPPER(cds.country_code)::bpchar
GROUP BY UPPER(cds.country_code)
ORDER BY moves_count DESC, code ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query country list: %w", err)
	}
	for i := range rows {
		rows[i].Code = strings.ToUpper(rows[i].Code)
		rows[i].Flag = countryFlag(rows[i].Code)
	}
	return rows, nil
}

func (s *Store) FetchWaypoint(ctx context.Context, waypointCode string) (WaypointDetails, error) {
	row := WaypointDetails{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	w.id,
	w.waypoint_code,
	w.source,
	UPPER(w.country) AS country,
	w.lat,
	w.lon,
	(
		SELECT COUNT(*)::bigint
		FROM geokrety.gk_geokrety_with_details AS g
		WHERE UPPER(g.waypoint) = UPPER(w.waypoint_code)
	) AS current_geokrety,
	(
		SELECT COUNT(DISTINCT gk_id)::bigint
		FROM stats.gk_cache_visits AS gcv
		WHERE gcv.waypoint_id = w.id
	) AS past_geokrety,
	(
		SELECT COALESCE(SUM(visit_count), 0)::bigint
		FROM stats.gk_cache_visits AS gcv
		WHERE gcv.waypoint_id = w.id
	) AS gk_visits,
	(
		SELECT COALESCE(SUM(visit_count), 0)::bigint
		FROM stats.user_cache_visits AS ucv
		WHERE ucv.waypoint_id = w.id
	) AS user_visits
FROM stats.waypoints AS w
WHERE UPPER(w.waypoint_code) = UPPER($1)
`, waypointCode); err != nil {
		return WaypointDetails{}, fmt.Errorf("query waypoint details: %w", err)
	}
	row.WaypointSummary = hydrateWaypointSummary(row.WaypointSummary)
	return row, nil
}

func (s *Store) FetchWaypointCurrentGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
WHERE UPPER(g.waypoint) = UPPER($1)
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $2 OFFSET $3
`, waypointCode, limit, offset); err != nil {
		return nil, fmt.Errorf("query waypoint current geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchWaypointPastGeokrety(ctx context.Context, waypointCode string, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	g.type,
	g.missing,
	g.owner AS owner_id,
	ou.username AS owner_username,
	g.holder AS holder_id,
	hu.username AS holder_username,
	NULL::varchar AS country,
	$1::varchar AS waypoint,
	NULL::double precision AS lat,
	NULL::double precision AS lon,
	g.loves_count,
	g.pictures_count,
	g.born_on_datetime AS born_at,
	g.updated_on_datetime AS last_move_at
FROM stats.gk_cache_visits AS gcv
INNER JOIN stats.waypoints AS w ON w.id = gcv.waypoint_id
INNER JOIN geokrety.gk_geokrety AS g ON g.id = gcv.gk_id
LEFT JOIN geokrety.gk_users AS ou ON ou.id = g.owner
LEFT JOIN geokrety.gk_users AS hu ON hu.id = g.holder
WHERE UPPER(w.waypoint_code) = UPPER($1)
ORDER BY gcv.last_visited_at DESC, g.id DESC
LIMIT $2 OFFSET $3
`, waypointCode, limit, offset); err != nil {
		return nil, fmt.Errorf("query waypoint past geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) SearchWaypoints(ctx context.Context, query string, limit, offset int) ([]WaypointSummary, error) {
	rows := []WaypointSummary{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	id,
	waypoint_code,
	source,
	UPPER(country) AS country,
	lat,
	lon
FROM stats.waypoints
WHERE waypoint_code ILIKE '%' || $1 || '%'
ORDER BY waypoint_code ASC
LIMIT $2 OFFSET $3
`, query, limit, offset); err != nil {
		return nil, fmt.Errorf("search waypoints: %w", err)
	}
	for i := range rows {
		rows[i] = hydrateWaypointSummary(rows[i])
	}
	return rows, nil
}

func (s *Store) FetchUserDetails(ctx context.Context, userID int64) (UserDetails, error) {
	row := UserDetails{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	u.id,
	u.username,
	u.joined_on_datetime AS joined_at,
	UPPER(u.home_country) AS home_country,
	u.avatar AS avatar_id,
	u.pictures_count,
	(
		SELECT COUNT(*)::bigint FROM geokrety.gk_geokrety AS g WHERE g.owner = u.id
	) AS owned_geokrety_count,
	(
		SELECT COUNT(*)::bigint FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS moves_count,
	(
		SELECT COUNT(DISTINCT m.geokret)::bigint FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS distinct_geokrety_count,
	(
		SELECT COUNT(*)::bigint FROM stats.user_countries AS uc WHERE uc.user_id = u.id
	) AS active_countries_count,
	(
		SELECT MAX(m.moved_on_datetime) FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS last_move_at
FROM geokrety.gk_users AS u
WHERE u.id = $1
`, userID); err != nil {
		return UserDetails{}, fmt.Errorf("query user details: %w", err)
	}
	if row.HomeCountry != nil {
		row.HomeCountryFlag = countryFlag(*row.HomeCountry)
	}
	return row, nil
}

func (s *Store) FetchUserOwnedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	return s.fetchUserGeokretyList(ctx, `g.owner = $1`, userID, limit, offset, "query user owned geokrety")
}

func (s *Store) FetchUserFoundGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT DISTINCT ON (g.id)
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	m.moved_on_datetime AS last_move_at
FROM geokrety.gk_moves AS m
INNER JOIN geokrety.gk_geokrety_with_details AS g ON g.id = m.geokret
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
WHERE m.author = $1
ORDER BY g.id, m.moved_on_datetime DESC, m.id DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user found geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchUserLovedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	l.created_on_datetime AS last_move_at
FROM geokrety.gk_loves AS l
INNER JOIN geokrety.gk_geokrety_with_details AS g ON g.id = l.geokret
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
WHERE l.user = $1
ORDER BY l.created_on_datetime DESC, l.id DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user loved geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchUserWatchedGeokrety(ctx context.Context, userID int64, limit, offset int) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	w.created_on_datetime AS last_move_at
FROM geokrety.gk_watched AS w
INNER JOIN geokrety.gk_geokrety_with_details AS g ON g.id = w.geokret
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
WHERE w.user = $1
ORDER BY w.created_on_datetime DESC, w.id DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user watched geokrety: %w", err)
	}
	return hydrateGeokretListItems(rows), nil
}

func (s *Store) FetchUserPictures(ctx context.Context, userID int64, limit, offset int) ([]PictureInfo, error) {
	rows := []PictureInfo{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT DISTINCT ON (p.id)
	p.id,
	p.type,
	p.filename,
	p.caption,
	p.key,
	p.geokret AS geokret_id,
	p.move AS move_id,
	p.user AS user_id,
	p.author AS author_id,
	u.username AS author_username,
	p.uploaded_on_datetime,
	p.created_on_datetime
FROM geokrety.gk_pictures AS p
LEFT JOIN geokrety.gk_users AS u ON u.id = p.author
WHERE p.user = $1 OR p.author = $1
ORDER BY p.id, p.created_on_datetime DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user pictures: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchUserCountries(ctx context.Context, userID int64, limit, offset int) ([]UserCountryVisit, error) {
	rows := []UserCountryVisit{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	UPPER(country_code) AS country_code,
	move_count,
	first_visit,
	last_visit
FROM stats.user_countries
WHERE user_id = $1
ORDER BY last_visit DESC, country_code ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user countries: %w", err)
	}
	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].CountryCode)
	}
	return rows, nil
}

func (s *Store) FetchUserWaypoints(ctx context.Context, userID int64, limit, offset int) ([]UserWaypointVisit, error) {
	rows := []UserWaypointVisit{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	w.waypoint_code,
	v.visit_count,
	v.first_visited_at,
	v.last_visited_at,
	UPPER(w.country) AS country,
	w.lat,
	w.lon
FROM stats.user_cache_visits AS v
INNER JOIN stats.waypoints AS w ON w.id = v.waypoint_id
WHERE v.user_id = $1
ORDER BY v.last_visited_at DESC, w.waypoint_code ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user waypoints: %w", err)
	}
	return hydrateWaypointVisits(rows), nil
}

func (s *Store) SearchUsers(ctx context.Context, query string, limit, offset int) ([]UserSearchResult, error) {
	rows := []UserSearchResult{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	u.id,
	u.username,
	u.joined_on_datetime AS joined_at,
	UPPER(u.home_country) AS home_country,
	(
		SELECT MAX(m.moved_on_datetime) FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS last_move_at
FROM geokrety.gk_users AS u
WHERE u.username ILIKE '%' || $1 || '%'
ORDER BY u.username ASC
LIMIT $2 OFFSET $3
`, query, limit, offset); err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchUserList(ctx context.Context, limit, offset int) ([]UserSearchResult, error) {
	rows := []UserSearchResult{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	u.id,
	u.username,
	u.joined_on_datetime AS joined_at,
	UPPER(u.home_country) AS home_country,
	(
		SELECT MAX(m.moved_on_datetime) FROM geokrety.gk_moves AS m WHERE m.author = u.id
	) AS last_move_at
FROM geokrety.gk_users AS u
ORDER BY u.username ASC, u.id ASC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query user list: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchUserStatsHeatmapDays(ctx context.Context, userID int64, limit, offset int) ([]DayHeatmapCell, error) {
	rows := []DayHeatmapCell{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	DATE(moved_on_datetime AT TIME ZONE 'UTC') AS day,
	COUNT(*)::bigint AS move_count
FROM geokrety.gk_moves
WHERE author = $1
GROUP BY DATE(moved_on_datetime AT TIME ZONE 'UTC')
ORDER BY day DESC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user heatmap days: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchUserStatsHeatmapHours(ctx context.Context, userID int64, limit, offset int) ([]HourHeatmapCell, error) {
	rows := []HourHeatmapCell{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	EXTRACT(ISODOW FROM moved_on_datetime AT TIME ZONE 'UTC')::int AS day_of_week,
	EXTRACT(HOUR FROM moved_on_datetime AT TIME ZONE 'UTC')::int AS hour_utc,
	COUNT(*)::bigint AS move_count
FROM geokrety.gk_moves
WHERE author = $1
GROUP BY EXTRACT(ISODOW FROM moved_on_datetime AT TIME ZONE 'UTC'), EXTRACT(HOUR FROM moved_on_datetime AT TIME ZONE 'UTC')
ORDER BY move_count DESC, day_of_week ASC, hour_utc ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user heatmap hours: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchUserStatsMapCountries(ctx context.Context, userID int64, limit, offset int) ([]CountryCount, error) {
	rows := []CountryCount{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	UPPER(country_code) AS country_code,
	move_count::bigint AS value,
	first_visit AS first_at
FROM stats.user_countries
WHERE user_id = $1
ORDER BY move_count DESC, country_code ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user map countries: %w", err)
	}
	for i := range rows {
		rows[i].Flag = countryFlag(rows[i].CountryCode)
	}
	return rows, nil
}

func (s *Store) FetchUserStatsContinentCoverage(ctx context.Context, userID int64, limit, offset int) ([]UserContinentCoverage, error) {
	rows := []UserContinentCoverage{}
	if err := s.db.SelectContext(ctx, &rows, `
WITH continent_names AS (
	SELECT
		continent_code,
		MIN(continent_name) AS continent_name
	FROM stats.continent_reference
	GROUP BY continent_code
), total_moves AS (
	SELECT COALESCE(SUM(moves), 0)::bigint AS total
	FROM stats.v_uc4_user_continent_coverage
	WHERE user_id = $1
)
SELECT
	v.user_id,
	COALESCE(u.username, 'unknown') AS username,
	v.continent_code,
	COALESCE(cn.continent_name, v.continent_code) AS continent_name,
	v.moves,
	CASE
		WHEN tm.total > 0 THEN ROUND((v.moves::numeric / tm.total::numeric), 4)::double precision
		ELSE 0
	END AS share
FROM stats.v_uc4_user_continent_coverage AS v
LEFT JOIN geokrety.gk_users AS u ON u.id = v.user_id
LEFT JOIN continent_names AS cn ON cn.continent_code = v.continent_code
CROSS JOIN total_moves AS tm
WHERE v.user_id = $1
ORDER BY v.moves DESC, v.continent_code ASC
LIMIT $2 OFFSET $3
`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("query user continent coverage: %w", err)
	}
	return rows, nil
}

func (s *Store) FetchPicture(ctx context.Context, pictureID int64) (PictureInfo, error) {
	row := PictureInfo{}
	if err := s.db.GetContext(ctx, &row, `
SELECT
	p.id,
	p.type,
	p.filename,
	p.caption,
	p.key,
	p.geokret AS geokret_id,
	p.move AS move_id,
	p.user AS user_id,
	p.author AS author_id,
	u.username AS author_username,
	p.uploaded_on_datetime,
	p.created_on_datetime
FROM geokrety.gk_pictures AS p
LEFT JOIN geokrety.gk_users AS u ON u.id = p.author
WHERE p.id = $1
`, pictureID); err != nil {
		return PictureInfo{}, fmt.Errorf("query picture details: %w", err)
	}
	return row, nil
}

func (s *Store) FetchPictureList(ctx context.Context, limit, offset int) ([]PictureInfo, error) {
	rows := []PictureInfo{}
	if err := s.db.SelectContext(ctx, &rows, `
SELECT
	p.id,
	p.type,
	p.filename,
	p.caption,
	p.key,
	p.geokret AS geokret_id,
	p.move AS move_id,
	p.user AS user_id,
	p.author AS author_id,
	u.username AS author_username,
	p.uploaded_on_datetime,
	p.created_on_datetime
FROM geokrety.gk_pictures AS p
LEFT JOIN geokrety.gk_users AS u ON u.id = p.author
ORDER BY p.created_on_datetime DESC, p.id DESC
LIMIT $1 OFFSET $2
`, limit, offset); err != nil {
		return nil, fmt.Errorf("query picture list: %w", err)
	}
	return rows, nil
}

func (s *Store) fetchUserGeokretyList(ctx context.Context, whereClause string, userID int64, limit, offset int, label string) ([]GeokretListItem, error) {
	rows := []GeokretListItem{}
	query := fmt.Sprintf(`
SELECT
	g.id,
	g.gkid,
	g.name,
	gg.type,
	gg.missing,
	gg.owner AS owner_id,
	NULLIF(g.owner_username, '') AS owner_username,
	gg.holder AS holder_id,
	hu.username AS holder_username,
	UPPER(g.country) AS country,
	UPPER(g.waypoint) AS waypoint,
	g.lat,
	g.lon,
	gg.loves_count,
	gg.pictures_count,
	gg.born_on_datetime AS born_at,
	g.moved_on_datetime AS last_move_at
FROM geokrety.gk_geokrety_with_details AS g
INNER JOIN geokrety.gk_geokrety AS gg ON gg.id = g.id
LEFT JOIN geokrety.gk_users AS hu ON hu.id = gg.holder
WHERE %s
ORDER BY g.moved_on_datetime DESC, g.id DESC
LIMIT $2 OFFSET $3
`, whereClause)
	if err := s.db.SelectContext(ctx, &rows, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return hydrateGeokretListItems(rows), nil
}
