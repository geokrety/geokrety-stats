# Feature: Country Leaderboard

**Status:** Complete  
**Date Created:** 2026-02-28  
**Last Updated:** 2026-02-28  
**Version:** 1.0

## Overview

The Country Leaderboard feature displays global rankings of countries by total points accumulated through GeoKrety movements. This provides a macro-level view of geographic competition and engagement, complementing the individual user leaderboard.

**Goal:** Show which regions/countries are most active and successful in the GeoKrety points system.

## Files Modified/Created

### Backend
- `leaderboard-api/internal/handlers/leaderboard.go` - Country ranking queries
- `leaderboard-api/internal/router.go` - New GET /api/users/countries endpoint

### Frontend
- `leaderboard-dashboard/src/views/CountryLeaderboardView.vue` - Country ranking table
- `leaderboard-dashboard/src/router/index.js` - New /countries route

### Database Views
- Query from `mv_user_stats` aggregated by country
- Uses materialized view refresh function

## API Endpoints

### GET /api/users/countries

Returns paginated list of countries ranked by total points.

**Method:** `GET`

**URL:** `http://localhost:8080/api/users/countries`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Maximum countries to return |
| `offset` | integer | 0 | Pagination offset |

**Response (200 OK):**
```json
[
  {
    "country": "France",
    "total_points": 1234567,
    "user_count": 42,
    "move_count": 5678,
    "rank": 1
  },
  {
    "country": "Germany",
    "total_points": 987654,
    "user_count": 38,
    "move_count": 4321,
    "rank": 2
  }
]
```

**Response Fields:**
- `country` (string) - Country name
- `total_points` (integer) - Sum of all points from users in this country
- `user_count` (integer) - Number of unique active users from this country
- `move_count` (integer) - Total number of GeoKrety moves by users in this country
- `rank` (integer) - Ranking position (1-based)

**Error Responses:**
- `400 Bad Request` - Invalid query parameters
- `500 Internal Server Error` - Database error

**Example Request:**
```bash
# Get top 50 countries
curl -s http://localhost:8080/api/users/countries | jq .

# Get top 10 countries with pagination
curl -s "http://localhost:8080/api/users/countries?limit=10&offset=0" | jq .

# Get countries 51-100
curl -s "http://localhost:8080/api/users/countries?limit=50&offset=50" | jq .
```

## Frontend Components

### CountryLeaderboardView.vue

**Location:** `leaderboard-dashboard/src/views/CountryLeaderboardView.vue`

**Purpose:** Display interactive table of country rankings

**Component Structure:**
```vue
<script setup>
// Imports
import { ref, onMounted, computed } from 'vue'
import { useLiveStats } from '../composables/useWebSocket.js'

// State
const countries = ref([])
const loading = ref(false)
const error = ref(null)
const { connected } = useLiveStats()

// Lifecycle
onMounted(() => {
  fetchCountries()
})

// Methods
const fetchCountries = async () => {
  // Fetch from /api/users/countries
}
</script>
```

**Features:**
- Ranked table display (top 50 countries)
- Color-coded medals for top 3 countries
  - 🥇 Gold for #1
  - 🥈 Silver for #2
  - 🥉 Bronze for #3
- Number formatting with thousand separators
- Responsive table layout (Bootstrap 5)
- Loading state indicator
- Error handling with user-friendly messages
- Live data updates via WebSocket

**Props:** None (fetches own data)

**Events:** None (trigger navigation on country click)

**Used Composables:**
- `useLiveStats()` - For connection status

## Testing Procedures

### API Testing with curl

**Test 1: Basic Request**
```bash
# Fetch top 50 countries
curl -s http://localhost:8080/api/users/countries | jq '.' | head -30
```

**Expected Result:**
- Array of objects
- Each has: country, total_points, user_count, move_count, rank
- Sorted by total_points descending
- First item should have rank=1

**Test 2: Pagination**
```bash
# Get first 10
curl -s "http://localhost:8080/api/users/countries?limit=10" | jq '.[].country'

# Get next 10
curl -s "http://localhost:8080/api/users/countries?limit=10&offset=10" | jq '.[].country'
```

**Expected Result:**
- First request: countries at positions 1-10
- Second request: countries at positions 11-20
- No overlap between requests

**Test 3: Count Verification**
```bash
# Verify count of results
curl -s "http://localhost:8080/api/users/countries?limit=50" | jq 'length'
```

**Expected Result:**
- Should return exactly 50 (or less if fewer countries exist)

**Test 4: Data Format**
```bash
# Check individual country data
curl -s http://localhost:8080/api/users/countries | jq '.[0]'
```

**Expected Result:**
```json
{
  "country": "France",
  "total_points": 1234567,
  "user_count": 42,
  "move_count": 5678,
  "rank": 1
}
```

### UI Testing with Gotenberg

**Test 1: Desktop View**
```bash
# Take desktop screenshot at 1280px width
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000/countries \
  --form width=1280 \
  --form height=1024 \
  -o /tmp/countries-desktop.png

file /tmp/countries-desktop.png
```

**Verification:**
- [ ] Table displays correctly
- [ ] Headers visible: Rank, Country, Points, Users, Moves
- [ ] Medal icons show for top 3
- [ ] Numbers formatted with commas
- [ ] All columns visible

**Test 2: Mobile View**
```bash
# Take mobile screenshot at 720px width
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000/countries \
  --form width=720 \
  --form height=2048 \
  -o /tmp/countries-mobile.png

file /tmp/countries-mobile.png
```

**Verification:**
- [ ] Table responsive (scrollable if needed)
- [ ] All data visible without horizontal scroll
- [ ] Readable on mobile screen
- [ ] Touch-friendly row heights

**Test 3: Data Display Accuracy**
```bash
# Compare API data with UI display
# Fetch top 3
curl -s "http://localhost:8080/api/users/countries?limit=3" | jq '.[] | "\(.rank): \(.country) - \(.total_points) points"'

# Take screenshot and verify numbers match
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000/countries \
  --form width=1280 \
  --form height=768 \
  -o /tmp/countries-verify.png
```

**Verification:**
- [ ] All countries from API visible in table
- [ ] Points match exactly
- [ ] User counts match
- [ ] Rankings correct

## Integration Testing

**Full Integration Test:**
```bash
# 1. Restart services
docker compose down
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d

# 2. Wait for ready
sleep 5

# 3. Test API
curl -s http://localhost:8080/api/users/countries | jq '.[0]'

# 4. Test UI
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000/countries \
  --form width=1280 --form height=1024 \
  -o /tmp/final-test.png

# 5. Verify logs clean
docker compose logs leaderboard-api | grep -i error
```

## Database

### Data Source

Query aggregates from `mv_user_stats` materialized view:

```sql
-- Conceptual query (actual implementation in handler)
SELECT 
  home_country as country,
  SUM(total_points) as total_points,
  COUNT(DISTINCT user_id) as user_count,
  SUM(total_moves) as move_count
FROM geokrety_stats.mv_user_stats
WHERE home_country IS NOT NULL
GROUP BY home_country
ORDER BY total_points DESC
LIMIT ? OFFSET ?
```

### Views Used
- `geokrety_stats.mv_user_stats` - User statistics including home_country

### Functions Called
- `geokrety_stats.refresh_leaderboard_views()` - Periodic materialized view refresh

### Performance Notes
- Materialized view provides fast aggregation
- No expensive joins at query time
- Refresh interval: typically 10 seconds
- Simple ORDER BY on pre-aggregated data

## WebSocket Integration

When deployed, country leaderboard may receive:

**Message Type:** `leaderboard_snapshot` or `leaderboard_update`
```json
{
  "type": "leaderboard_snapshot",
  "payload": [
    {
      "country": "France",
      "total_points": 1234567,
      ...
    }
  ]
}
```

**Behavior:**
- CountryLeaderboardView doesn't directly listen to WebSocket
- Uses standard REST API
- Could be enhanced to auto-refresh on leaderboard_update

## Deployment Notes

### Prerequisites
- Database must have `mv_user_stats` view with `home_country` column
- Materialized views must be refreshed regularly
- API endpoint registered in router

### Build Process
```bash
cd /home/kumy/GIT/geokrety-points-system
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d
```

### Verification After Deploy
```bash
# Check API responds
curl -s http://localhost:8080/api/users/countries | jq '.[0]'

# Check UI loads
curl -s http://localhost:3000/countries | grep -i "leaderboard"

# Check for errors
docker compose logs leaderboard-api | grep -i error
docker compose logs leaderboard-dashboard | grep -i error
```

## Known Issues / Limitations

- **No user geo-location validation** - Uses home_country from user profile (not validated)
- **No real-time updates** - Country leaderboard is REST-only, not WebSocket streamed
- **Country name standardization** - Depends on consistent country naming in user profiles
- **No filtering** - Cannot filter by region or continent
- **No historical data** - Shows current snapshot only

## Future Enhancements

1. **Search and Filter**
   - Filter by continent/region
   - Search specific country

2. **Leaderboard Per Country**
   - Show top users within each country
   - Country-specific statistics

3. **Historical Tracking**
   - Chart country ranking progression
   - Compare across dates

4. **Export**
   - Download as CSV
   - Export country statistics

5. **Real-time Updates**
   - WebSocket streaming of country leaderboard
   - Auto-refresh on broadcast

6. **Analytics**
   - Country growth trends
   - Engagement metrics by region

## Related Features

- [breakdown-charts.md](breakdown-charts.md) - Statistics and visualizations
- [websocket-user-count.md](websocket-user-count.md) - Live user count

---

**Last Updated:** 2026-02-28  
**Version:** 1.0  
**Maintainer:** Development Team
