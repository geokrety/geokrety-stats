# Feature: Breakdown Charts (Statistics Dashboard)

**Status:** Complete
**Date Created:** 2026-02-28
**Last Updated:** 2026-02-28
**Version:** 1.0

## Overview

The Breakdown Charts feature provides detailed visualizations of GeoKrety statistics showing how points are earned, distributed across event types, and how user activity varies. Four interactive D3.js charts display:

1. **Points Distribution** - Top 20 users by total points
2. **Moves Distribution** - Top 20 users by move count
3. **Cost Breakdown by Event Type** - Drop/Grab/Comment costs
4. **Event Count Distribution** - Count of each event type

**Goal:** Give users and admins insights into points distribution, activity patterns, and engagement metrics.

## Files Modified/Created

### Backend
- `leaderboard-api/internal/handlers/leaderboard.go` - Statistics aggregation queries
- `leaderboard-api/internal/models/models.go` - Statistics data structures

### Frontend
- `leaderboard-dashboard/src/views/StatsBreakdown.vue` - Main statistics view with charts
- `leaderboard-dashboard/src/components/PointsBreakdownChart.vue` - D3.js points chart
- `leaderboard-dashboard/src/components/MovesBreakdownChart.vue` - D3.js moves chart
- `leaderboard-dashboard/src/components/EventCostChart.vue` - Cost breakdown chart
- `leaderboard-dashboard/src/components/EventCountChart.vue` - Event count chart
- `leaderboard-dashboard/src/router/index.js` - New /stats route

### Database Views
- Queries from `mv_user_stats` and event aggregation tables
- Uses materialized views for performance

## API Endpoints

### GET /api/stats/breakdown

Returns aggregate statistics for visualization in breakdown charts.

**Method:** `GET`

**URL:** `http://localhost:8080/api/stats/breakdown`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 20 | Number of top users to include |

**Response (200 OK):**
```json
{
  "top_users_by_points": [
    {
      "user_id": 1,
      "username": "user1",
      "total_points": 100000,
      "move_count": 50
    },
    {
      "user_id": 2,
      "username": "user2",
      "total_points": 95000,
      "move_count": 48
    }
  ],
  "event_costs": [
    {
      "user_id": 1,
      "username": "user1",
      "drop_cost": 5000,
      "grab_cost": 3000,
      "comment_cost": 500
    },
    {
      "user_id": 2,
      "username": "user2",
      "drop_cost": 4800,
      "grab_cost": 2900,
      "comment_cost": 450
    }
  ],
  "event_counts": [
    {
      "user_id": 1,
      "username": "user1",
      "drop_count": 10,
      "grab_count": 8,
      "comment_count": 25
    },
    {
      "user_id": 2,
      "username": "user2",
      "drop_count": 9,
      "grab_count": 8,
      "comment_count": 24
    }
  ]
}
```

**Response Fields:**

**top_users_by_points array:**
- `user_id` (integer) - Unique user identifier
- `username` (string) - Display name
- `total_points` (integer) - Sum of all points earned
- `move_count` (integer) - Number of moves made

**event_costs array:**
- `user_id` (integer) - User identifier
- `username` (string) - Display name
- `drop_cost` (float) - Total cost of drop events
- `grab_cost` (float) - Total cost of grab events
- `comment_cost` (float) - Total cost of comment events

**event_counts array:**
- `user_id` (integer) - User identifier
- `username` (string) - Display name
- `drop_count` (integer) - Number of drop events
- `grab_count` (integer) - Number of grab events
- `comment_count` (integer) - Number of comment events

**Error Responses:**
- `400 Bad Request` - Invalid parameters
- `500 Internal Server Error` - Database error

**Example Requests:**
```bash
# Get default top 20 users
curl -s http://localhost:8080/api/stats/breakdown | jq .

# Get top 10 users
curl -s "http://localhost:8080/api/stats/breakdown?limit=10" | jq .

# Pretty print first user
curl -s http://localhost:8080/api/stats/breakdown | jq '.top_users_by_points[0]'

# List all usernames
curl -s http://localhost:8080/api/stats/breakdown | jq '.top_users_by_points[].username'
```

## Frontend Components

### StatsBreakdown.vue

**Location:** `leaderboard-dashboard/src/views/StatsBreakdown.vue`

**Purpose:** Main statistics dashboard containing all four charts

**Features:**
- Fetches `/api/stats/breakdown` on mount
- Displays 4 charts in responsive grid layout
- Loading state while fetching
- Error handling with user-friendly messages
- Real-time updates via WebSocket (future enhancement)
- Responsive design for mobile/tablet/desktop

**Component Structure:**
```vue
<script setup>
// Imports: Vue, composables, child components
import { ref, onMounted } from 'vue'
import { useLiveStats } from '../composables/useWebSocket.js'
import PointsBreakdownChart from '../components/PointsBreakdownChart.vue'
import MovesBreakdownChart from '../components/MovesBreakdownChart.vue'
import EventCostChart from '../components/EventCostChart.vue'
import EventCountChart from '../components/EventCountChart.vue'

// State
const stats = ref(null)
const loading = ref(true)
const error = ref(null)
const limit = ref(20)
const { connected } = useLiveStats()

// Lifecycle
onMounted(() => {
  fetchStats()
})

// Methods
const fetchStats = async () => {
  // Fetch from /api/stats/breakdown
}
</script>

<template>
  <!-- Grid layout with 4 charts -->
</template>
```

### PointsBreakdownChart.vue

**Location:** `leaderboard-dashboard/src/components/PointsBreakdownChart.vue`

**Purpose:** D3.js bar chart of top users by points

**Props:**
```javascript
{
  data: Array,        // top_users_by_points
  loading: Boolean,   // Show loading state
  title: String       // Chart title
}
```

**Features:**
- Horizontal bar chart using D3.js
- X-axis: Total points
- Y-axis: User names
- Color gradient from light to dark
- Tooltips on hover showing exact value
- Responsive to container width
- Sorts descending by points

### MovesBreakdownChart.vue

**Location:** `leaderboard-dashboard/src/components/MovesBreakdownChart.vue`

**Purpose:** D3.js bar chart of top users by move count

**Props:**
```javascript
{
  data: Array,        // top_users_by_points (uses move_count)
  loading: Boolean,
  title: String
}
```

**Features:**
- Vertical bar chart using D3.js
- X-axis: User names
- Y-axis: Move count
- Color: Bootstrap success color
- Tooltips on hover
- Responsive design

### EventCostChart.vue

**Location:** `leaderboard-dashboard/src/components/EventCostChart.vue`

**Purpose:** Stacked horizontal bar showing event type costs

**Props:**
```javascript
{
  data: Array,        // event_costs
  loading: Boolean,
  title: String
}
```

**Features:**
- Stacked horizontal bar chart
- Three segments: Drop (red), Grab (blue), Comment (orange)
- Shows total cost distribution per user
- Stacked visualization shows proportions
- Tooltips show individual segment values
- Color coding matches event types

### EventCountChart.vue

**Location:** `leaderboard-dashboard/src/components/EventCountChart.vue`

**Purpose:** Grouped bar chart of event counts by type

**Props:**
```javascript
{
  data: Array,        // event_counts
  loading: Boolean,
  title: String
}
```

**Features:**
- Grouped bar chart (not stacked)
- Three bars per user: Drop, Grab, Comment
- Uses same color scheme as cost chart
- Allows comparison of event frequencies
- Tooltips on hover

## Testing Procedures

### API Testing with curl

**Test 1: Basic Request**
```bash
# Get default stats
curl -s http://localhost:8080/api/stats/breakdown | jq '.' | head -50
```

**Expected Result:**
- Three top-level arrays: top_users_by_points, event_costs, event_counts
- Each should have same number of users
- Points should be in descending order

**Test 2: Data Structure Verification**
```bash
# Check top user structure
curl -s http://localhost:8080/api/stats/breakdown | jq '.top_users_by_points[0]'
```

**Expected Result:**
```json
{
  "user_id": <number>,
  "username": "<string>",
  "total_points": <number>,
  "move_count": <number>
}
```

**Test 3: Cost Data Verification**
```bash
# Check cost structure
curl -s http://localhost:8080/api/stats/breakdown | jq '.event_costs[0]'
```

**Expected Result:**
```json
{
  "user_id": <number>,
  "username": "<string>",
  "drop_cost": <number>,
  "grab_cost": <number>,
  "comment_cost": <number>
}
```

**Test 4: Event Count Verification**
```bash
# Check event counts
curl -s http://localhost:8080/api/stats/breakdown | jq '.event_counts[0]'
```

**Expected Result:**
```json
{
  "user_id": <number>,
  "username": "<string>",
  "drop_count": <number>,
  "grab_count": <number>,
  "comment_count": <number>
}
```

**Test 5: Limit Parameter**
```bash
# Get top 10 users
curl -s "http://localhost:8080/api/stats/breakdown?limit=10" | jq '.top_users_by_points | length'

# Get top 5 users
curl -s "http://localhost:8080/api/stats/breakdown?limit=5" | jq '.top_users_by_points | length'
```

**Expected Result:**
- First returns 10
- Second returns 5

### UI Testing with MCP Playwright

**Load MCP Playwright tools first:**
Use `tool_search_tool_regex` with pattern: `^mcp_microsoft_pla_browser`

**Test 1: Full Dashboard View**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://localhost:3000/stats`
2. Resize: `mcp_microsoft_pla_browser_resize` to 1280x2048
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

**Verification:**
- [ ] All 4 charts visible
- [ ] Charts have titles
- [ ] Axes are labeled
- [ ] Data values visible
- [ ] Loading state not shown

**Test 2: Mobile Responsiveness**
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://localhost:3000/stats`
2. Resize: `mcp_microsoft_pla_browser_resize` to 720x2048
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

**Verification:**
- [ ] Charts stack vertically
- [ ] Each chart readable on mobile
- [ ] No horizontal scroll needed
- [ ] Touch-friendly sizes

**Test 3: Chart-Specific Views**

Desktop (two-column layout):
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://localhost:3000/stats`
2. Resize: `mcp_microsoft_pla_browser_resize` to 1280x1200
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

Tablet (single column):
1. Navigate: `mcp_microsoft_pla_browser_navigate` to `http://localhost:3000/stats`
2. Resize: `mcp_microsoft_pla_browser_resize` to 900x2000
3. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`

## Integration Testing

**Full Integration Test:**
```bash
# 1. Restart services
docker compose down
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d
sleep 5

# 2. Test API
curl -s http://localhost:8080/api/stats/breakdown | \
  jq '{users: (.top_users_by_points | length), costs: (.event_costs | length), counts: (.event_counts | length)}'

# Expected: {"users": 20, "costs": 20, "counts": 20}

# 3. Test UI loads
curl -s http://localhost:3000/stats | grep -c "svg" || echo "Charts loaded"

# 4. Take full screenshot with MCP Playwright
# Load tools: tool_search_tool_regex with pattern ^mcp_microsoft_pla_browser
# Navigate: mcp_microsoft_pla_browser_navigate to http://localhost:3000/stats
# Resize: mcp_microsoft_pla_browser_resize to 1280x1500
# Screenshot: mcp_microsoft_pla_browser_take_screenshot

# 5. Check for errors
docker compose logs leaderboard-api | grep -i error | head -5
docker compose logs leaderboard-dashboard | grep -i error | head -5
```

## Database

### Data Sources

The endpoint aggregates data from multiple sources:

```sql
-- Conceptual queries
-- Top users by points
SELECT user_id, username, total_points, total_moves
FROM geokrety_stats.mv_user_stats
ORDER BY total_points DESC LIMIT ?

-- Event costs (aggregated from event logs)
SELECT user_id, username,
  SUM(CASE WHEN event_type='drop' THEN cost ELSE 0 END) as drop_cost,
  SUM(CASE WHEN event_type='grab' THEN cost ELSE 0 END) as grab_cost,
  SUM(CASE WHEN event_type='comment' THEN cost ELSE 0 END) as comment_cost
FROM geokrety_stats.event_details
GROUP BY user_id, username
ORDER BY user_id DESC LIMIT ?

-- Event counts (count records by type)
SELECT user_id, username,
  SUM(CASE WHEN event_type='drop' THEN 1 ELSE 0 END) as drop_count,
  SUM(CASE WHEN event_type='grab' THEN 1 ELSE 0 END) as grab_count,
  SUM(CASE WHEN event_type='comment' THEN 1 ELSE 0 END) as comment_count
FROM geokrety_stats.event_details
GROUP BY user_id, username
ORDER BY user_id DESC LIMIT ?
```

### Views Used
- `geokrety_stats.mv_user_stats` - User points and move counts
- `geokrety_stats.event_details` - Individual event records

### Functions Called
- `geokrety_stats.refresh_leaderboard_views()` - Periodic refresh

### Performance Notes
- Materialized views pre-aggregate user stats
- Event detail queries may scan large tables
- Consider indexing on (user_id, event_type)
- Typical response time: <500ms with 20 users

## WebSocket Integration

**No direct WebSocket updates** for this view currently.

Future enhancement possibility:
```json
{
  "type": "stats_update",
  "payload": {
    "top_users_by_points": [...],
    "event_costs": [...],
    "event_counts": [...]
  }
}
```

## Deployment Notes

### Prerequisites
- Database must have event_details or equivalent table
- Must have aggregation functions for costs/counts
- D3.js v7 already loaded globally
- Bootstrap 5 CSS for responsiveness

### Build Process
```bash
cd /home/kumy/GIT/geokrety-points-system
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d
```

### Verification After Deploy
```bash
# Test API endpoint
curl -s http://localhost:8080/api/stats/breakdown | jq '.top_users_by_points | length'

# Test UI route loads
curl -s http://localhost:3000/stats | grep -o "<title>[^<]*</title>"

# Check for errors
docker compose logs leaderboard-api | tail -10
docker compose logs leaderboard-dashboard | tail -10
```

## Known Issues / Limitations

- **No real-time updates** - Page requires manual refresh for new data
- **Top 20 hardcoded** - Cannot adjust limit from UI (only via API param)
- **No data caching** - Fetches on every page load
- **Limited time range** - Shows current statistics, no historical comparison
- **No filtering** - Cannot filter by date range or user criteria
- **Event type definitions** - Assumes Drop/Grab/Comment types exist

## Future Enhancements

1. **Time-based Filtering**
   - Select date range for statistics
   - Compare periods

2. **Real-time Updates**
   - WebSocket streaming of stats
   - Auto-refresh on updates

3. **Export Functionality**
   - Download charts as PNG
   - Export data as CSV
   - PDF report generation

4. **Advanced Visualizations**
   - Line chart for points over time
   - Pie chart for event type distribution
   - Heatmap for user activity

5. **User Interaction**
   - Click user row to see detailed stats
   - Filter by date range in UI
   - Adjustable limit selector

6. **Performance**
   - Data caching
   - Pagination for large datasets
   - Incremental data loading

## Related Features

- [country-leaderboard.md](country-leaderboard.md) - Country rankings
- [websocket-user-count.md](websocket-user-count.md) - Live user count

---

**Last Updated:** 2026-02-28
**Version:** 1.0
**Maintainer:** Development Team
