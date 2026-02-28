# GeoKrety Leaderboard Dashboard - New Features

This document describes the new features implemented in the GeoKrety Leaderboard Dashboard during the most recent enhancement phase.

## Overview

The dashboard has been significantly enhanced with three major new features:
1. **Country Leaderboard** - Global country rankings by total points
2. **Breakdown Charts** - Detailed visualizations of points, moves, costs, and event distributions
3. **WebSocket Connected Users Count** - Real-time display of active dashboard users

## Feature Details

### 1. Country Leaderboard Page

#### Location
- **Route**: `/countries`
- **Component**: `leaderboard-dashboard/src/views/CountriesLeaderboard.vue`

#### Features
- Ranks countries by total points accumulated (descending)
- Displays top 50 countries with:
  - Country rank (1-50)
  - Country name with flag emoji
  - Total points (formatted with thousand separators)
  - Number of active users from that country
  - Move count per country

#### Data Source
- Fetches from REST API endpoint: `GET /api/users/countries`
- Returns paginated country statistics sorted by total points

#### User Experience
- Clean, responsive table layout
- Color-coded ranks (gold for #1, silver for #2, bronze for #3)
- Automatic number formatting (e.g., "1,234,567")
- Tooltip indicators for user/move counts
- Mobile-friendly responsive design

#### Technical Implementation
- Vue 3 Composition API with `<script setup>`
- Reactive data binding for real-time updates
- Bootstrap 5 table styling with custom enhancements
- Error handling with user-friendly messages

---

### 2. Breakdown Charts (Statistics Page)

#### Location
- **Route**: `/stats`
- **Component**: `leaderboard-dashboard/src/views/StatsBreakdown.vue`

#### Charts Included

##### Chart 1: Points Distribution
- **Type**: Horizontal bar chart
- **Data**: Top 20 users by total points
- **Visualization**: D3.js bar chart
- **Interactivity**: Hover tooltips showing exact values
- **Color Scheme**: Bootstrap primary with gradients

##### Chart 2: Moves Distribution
- **Type**: Vertical bar chart
- **Data**: Top 20 users by move count
- **Visualization**: D3.js bar chart
- **Focus**: Shows movement activity levels
- **Color Scheme**: Bootstrap success with gradients

##### Chart 3: Cost Breakdown by Event Type
- **Type**: Horizontal stacked bar chart
- **Data**: Three event types: Drop, Grab, Comment
- **Visualization**: Shows proportional cost contribution per user
- **Metric**: Total event costs by type for top 20 users
- **Color Scheme**: Red (Drop), Blue (Grab), Orange (Comment)

##### Chart 4: Event Count Distribution by Type
- **Type**: Grouped bar chart
- **Data**: Count of Drop, Grab, and Comment events
- **Scope**: Top 20 users
- **Visualization**: Side-by-side bars for comparison
- **Color Scheme**: Red (Drop), Blue (Grab), Orange (Comment)

#### Data Source
- **Endpoint**: `GET /api/stats/breakdown`
- **Response Format**: JSON with aggregated statistics

#### Technical Implementation
- **D3.js v7**: SVG-based visualization library
- **Responsive Design**: Charts scale with container width
- **Animation**: Smooth transitions on data updates
- **Accessibility**: ARIA labels and semantic HTML
- **Performance**: Debounced resize handling

#### Features
- Real-time data updates via WebSocket
- Responsive charts that adapt to screen size
- Proper axis labels and legends
- Value tooltips on hover
- Mobile-friendly collapse mode

---

### 3. WebSocket Connected Users Count

#### Location
- **Display**: Application footer
- **Component**: `leaderboard-dashboard/src/App.vue` (footer section)

#### Features
- **Real-time Count**: Shows current number of connected users
- **Status Indicator**:
  - Green when WebSocket is connected with user count
  - Gray when offline with fallback message
- **Automatic Updates**: Refreshes every 10 seconds via broadcaster
- **Plural Handling**: Grammatically correct "user"/"users" display

#### Message Format
```json
{
  "type": "connected_users",
  "payload": {
    "count": 5
  }
}
```

#### Data Flow
1. **Backend** (`leaderboard-api`):
   - `StartBroadcaster()` in `internal/handlers/websocket.go`
   - Sends `connected_users` message every broadcast interval
   - Uses `hub.ClientCount()` to get active connections

2. **Frontend** (`leaderboard-dashboard`):
   - WebSocket composable `useWebSocket.js` receives message
   - Updates reactive `connectedUsers` ref
   - App.vue footer displays count with live indicator

#### Technical Implementation
- **WebSocket Protocol**: Standard RFC 6455
- **Message Type**: "connected_users"
- **Update Frequency**: Every broadcast interval (typically 10 seconds)
- **Storage**: Global reactive ref in composable
- **Display**: Conditional rendering with Bootstrap icons

#### Styling
- Uses Bootstrap Icons (`bi-people-fill`)
- Color-coded status (text-success when live, text-secondary when offline)
- Integrated seamlessly into existing footer

---

## API Endpoints

### New Endpoints Added

#### 1. Country Leaderboard
```
GET /api/users/countries
```
**Query Parameters:**
- `limit` (optional, default: 50)
- `offset` (optional, default: 0)

**Response:**
```json
[
  {
    "country": "France",
    "total_points": 1234567,
    "user_count": 42,
    "move_count": 5678
  },
  ...
]
```

#### 2. Statistics Breakdown
```
GET /api/stats/breakdown
```
**Query Parameters:**
- `limit` (optional, default: 20)

**Response:**
```json
{
  "top_users_by_points": [
    {
      "user_id": 1,
      "username": "user1",
      "total_points": 100000,
      "move_count": 50
    },
    ...
  ],
  "event_costs": [
    {
      "user_id": 1,
      "username": "user1",
      "drop_cost": 5000,
      "grab_cost": 3000,
      "comment_cost": 500
    },
    ...
  ],
  "event_counts": [
    {
      "user_id": 1,
      "username": "user1",
      "drop_count": 10,
      "grab_count": 8,
      "comment_count": 25
    },
    ...
  ]
}
```

#### 3. WebSocket Live Updates
```
WebSocket /ws
```
**Message Types:**
- `leaderboard_snapshot` - Initial leaderboard data
- `leaderboard_update` - Periodic leaderboard updates
- `global_stats` - Global statistics
- `connected_users` - Active user count

---

## Database Views and Functions

### Views Used
- `mv_user_stats` - User statistics (points, moves, etc.)
- `mv_country_stats` - Country leaderboard aggregates
- `mv_event_details` - Event breakdown by type

### Functions Called
- `geokrety_stats.refresh_leaderboard_views()` - Refreshes materialized views
- Custom aggregation functions for statistics

---

## Frontend Components

### Views
- **[CountriesLeaderboard.vue](leaderboard-dashboard/src/views/CountriesLeaderboard.vue)** - Country rankings
- **[StatsBreakdown.vue](leaderboard-dashboard/src/views/StatsBreakdown.vue)** - Detailed breakdown charts
- **[App.vue](leaderboard-dashboard/src/App.vue)** - Main app with footer user count

### Composables
- **[useWebSocket.js](leaderboard-dashboard/src/composables/useWebSocket.js)** - WebSocket management and state

### Supporting Files
- Router configuration updated to include `/countries` and `/stats` routes

---

## Performance Considerations

### Frontend
- Charts use D3.js for efficient SVG rendering
- Debounced resize handlers prevent excessive re-renders
- Virtual scrolling not needed for table sizes (<50 rows)
- CSS Grid/Flexbox for responsive layouts

### Backend
- Materialized views for fast data aggregation
- WebSocket broadcasts with backpressure handling
- Limited client count checks before broadcasting
- Message queue with overflow protection

### Network
- Gzip compression on API responses
- WebSocket persistent connections reduce round-trips
- Batch updates every 10 seconds
- Small message payloads (~1-2KB per update)

---

## Browser Compatibility

- Chrome/Chromium 90+
- Firefox 88+
- Safari 14+
- Edge 90+
- Mobile browsers: iOS Safari 14+, Chrome Mobile 90+

---

## Testing Recommendations

### Unit Tests
- Component rendering tests for new views
- WebSocket message type parsing
- API response data transformation

### Integration Tests
- End-to-end country leaderboard flow
- Real-time stats update synchronization
- WebSocket connection lifecycle

### Manual Testing Checklist
- [ ] Navigate to /countries and verify table displays
- [ ] Verify country rankings are sorted correctly
- [ ] Check responsive design on mobile
- [ ] Test /stats page with all 4 charts visible
- [ ] Hover over charts to see tooltips
- [ ] Verify WebSocket user count updates
- [ ] Test offline/online status indicator in footer
- [ ] Verify data refreshes after API updates

---

## Future Enhancement Ideas

1. **Search and Filter**
   - Filter countries by region/continent
   - Search specific users in breakdown charts

2. **Leaderboard Per Country**
   - Show top users within each country
   - Country-specific statistics

3. **Historical Data**
   - Charts showing points progression over time
   - Compare country rankings across dates

4. **Export Functionality**
   - Download leaderboard as CSV
   - Export charts as PNG

5. **Notifications**
   - Toast notifications for connection status
   - Achievement notifications

6. **Advanced Analytics**
   - User activity heatmaps
   - Movement patterns visualization

---

## Deployment

### Prerequisites
- Docker and Docker Compose
- PostgreSQL 12+ with GeoKrety schema
- Node.js 18+ (for development)
- Go 1.19+ (for API development)

### Build and Deploy
```bash
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d
```

### Verification
- Frontend: http://localhost:3000
- API: http://localhost:8080
- Check logs: `docker compose logs -f leaderboard-dashboard`

---

## Troubleshooting

### Charts Not Rendering
- Check browser console for D3.js errors
- Verify API `/api/stats/breakdown` returns valid data
- Check container logs: `docker compose logs leaderboard-dashboard`

### WebSocket Disconnection Issues
- Check network tab in DevTools
- Verify API is running: `docker compose ps`
- Check API logs for WebSocket errors
- Look for firewall/proxy issues blocking WebSocket upgrade

### Country Data Missing
- Verify `mv_country_stats` exists in database
- Confirm materialized views are refreshed
- Check database connection in API logs

---

## Summary

These three new features significantly enhance the GeoKrety Leaderboard Dashboard:

- **Country Leaderboard** provides global competitive rankings
- **Breakdown Charts** offer detailed performance insights
- **Connected Users Count** shows real-time community engagement

All features are production-ready, fully responsive, and integrated with the existing WebSocket infrastructure for live updates.
