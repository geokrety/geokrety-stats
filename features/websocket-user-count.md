# Feature: WebSocket Connected Users Count

**Status:** Complete  
**Date Created:** 2026-02-28  
**Last Updated:** 2026-02-28  
**Version:** 1.0

## Overview

The WebSocket Connected Users feature displays real-time count of actively connected dashboard users in the application footer. This provides live engagement metrics showing how many people are currently using the GeoKrety Leaderboard.

**Goal:** Show real-time dashboard engagement with a live user count indicator.

## Files Modified/Created

### Backend
- `leaderboard-api/internal/handlers/websocket.go` - WebSocket message broadcasting
- `leaderboard-api/internal/websocket/hub.go` - Client count tracking

### Frontend
- `leaderboard-dashboard/src/composables/useWebSocket.js` - WebSocket connection management
- `leaderboard-dashboard/src/App.vue` - Footer with user count display

## API Endpoints

### WebSocket: GET /ws

Establishes persistent WebSocket connection for real-time updates.

**Protocol:** WebSocket (RFC 6455)

**URL:** `ws://localhost:8080/ws` (or `wss://` for HTTPS)

**Connection Lifecycle:**
1. Client connects to `/ws` endpoint
2. Server registers client in hub
3. Client receives periodic broadcasts
4. Client disconnects (on close or timeout)
5. Server unregisters client

**Message Format:**

All messages use JSON format:
```json
{
  "type": "message_type",
  "payload": {...}
}
```

## WebSocket Messages

### connected_users

Broadcasts current number of connected dashboard users every 10 seconds.

**Message Type:** `connected_users`

**Payload:**
```json
{
  "type": "connected_users",
  "payload": {
    "count": 5
  }
}
```

**Payload Fields:**
- `count` (integer) - Number of currently connected WebSocket clients

**Broadcast Frequency:** Every 10 seconds (same as main leaderboard refresh interval)

**Example Message:**
```json
{
  "type": "connected_users",
  "payload": {
    "count": 3
  }
}
```

### leaderboard_snapshot

Initial leaderboard data sent immediately upon connection.

**Message Type:** `leaderboard_snapshot`

**Payload:** Array of top 10 users (see [country-leaderboard.md](country-leaderboard.md))

### leaderboard_update

Periodic leaderboard updates broadcast to all connected clients.

**Message Type:** `leaderboard_update`

**Payload:** Array of top 10 users with current data

### global_stats

Global statistics update broadcast periodically.

**Message Type:** `global_stats`

**Payload:**
```json
{
  "total_points": 1000000,
  "total_moves": 50000,
  "total_geokreties": 500,
  "active_users": 250
}
```

## Frontend Composables

### useWebSocket.js

**Location:** `leaderboard-dashboard/src/composables/useWebSocket.js`

**Purpose:** Manage WebSocket connection and expose reactive state to components

**Exports:**

#### useLiveStats()
Main composable for reading and managing WebSocket state.

**Returns:**
```javascript
{
  connected: Ref<Boolean>,      // true if WebSocket connected
  leaderboard: Ref<Array>,      // Current leaderboard data
  stats: Ref<Object>,           // Global statistics
  connectedUsers: Ref<Number>   // Number of connected users
}
```

**Usage:**
```vue
<script setup>
import { useLiveStats } from '../composables/useWebSocket.js'

const { connected, stats, connectedUsers } = useLiveStats()
</script>

<template>
  <div v-if="connected" class="text-success">
    <i class="bi bi-wifi"></i> {{ connectedUsers }} users online
  </div>
  <div v-else class="text-secondary">
    <i class="bi bi-wifi-off"></i> Offline
  </div>
</template>
```

**Features:**
- Automatically connects on mount
- Automatically reconnects if disconnected (5-second retry)
- Reactive refs for template binding
- Handles all message types
- Global state shared across components

#### useLeaderboardLive()
Simplified composable for read-only leaderboard access (deprecated).

**Returns:**
```javascript
{
  connected: Ref<Boolean>,
  leaderboard: Ref<Array>
}
```

## Frontend Components

### App.vue Footer

**Location:** `leaderboard-dashboard/src/App.vue`

**Component Section:** Footer (bottom of page)

**Current Implementation:**
```vue
<footer class="bg-dark text-secondary text-center py-2 small mt-4">
  GeoKrety Points System &mdash; 
  <span v-if="connected" class="text-success">
    <i class="bi bi-people-fill me-1"></i>{{ connectedUsers }} user{{ connectedUsers !== 1 ? 's' : '' }} online
  </span>
  <span v-else class="text-secondary">
    Data refreshes when connection restored
  </span>
</footer>
```

**Features:**
- Displays user count when connected
- Shows offline status when disconnected
- Proper plural handling ("1 user" vs "5 users")
- Bootstrap Icons for visual indicator
- Color-coded (green when connected, gray when offline)
- Responsive footer styling

**State Props:**
- `connected` (Boolean) - WebSocket connection status
- `connectedUsers` (Number) - Count of connected users

## Testing Procedures

### WebSocket Connection Testing

**Test 1: Connection Establishment**
```bash
# Open WebSocket connection with wscat (if available)
wscat -c ws://localhost:8080/ws

# Or use Python
python3 << 'PYTHON'
import websocket
import json

ws = websocket.create_connection("ws://localhost:8080/ws")
for i in range(3):
    msg = ws.recv()
    print(json.dumps(json.loads(msg), indent=2))
    
ws.close()
PYTHON
```

**Expected Result:**
- Connection successful (no error)
- Receives messages within 2 seconds
- Messages are valid JSON
- At least one message type: `connected_users`

**Test 2: User Count Message**
```bash
# Connect and capture messages
python3 << 'PYTHON'
import websocket
import json
import time

ws = websocket.create_connection("ws://localhost:8080/ws")

# Look for connected_users message
start = time.time()
while time.time() - start < 15:
    try:
        msg = ws.recv()
        data = json.loads(msg)
        if data.get('type') == 'connected_users':
            count = data.get('payload', {}).get('count', 'N/A')
            print(f"Connected users: {count}")
            break
    except:
        continue
        
ws.close()
PYTHON
```

**Expected Result:**
- Receives `connected_users` message
- Message has payload.count with integer value
- Count is >= 1 (at least this connection)

**Test 3: Message Frequency**
```bash
# Count messages received in 15 seconds
python3 << 'PYTHON'
import websocket
import json
import time

ws = websocket.create_connection("ws://localhost:8080/ws")
start = time.time()
count = 0

while time.time() - start < 15:
    try:
        msg = ws.recv()
        count += 1
    except:
        break
        
ws.close()
print(f"Received {count} messages in 15 seconds")
print(f"Message rate: ~{count/15:.1f} messages/second")  # Should be ~1-2 messages/second
PYTHON
```

**Expected Result:**
- Receives multiple messages
- ~1-2 messages per second (or per broadcast interval)
- No connection errors

### API Testing with curl

**Test 1: Verify API is WebSocket-enabled**
```bash
# Check API health
curl -s http://localhost:8080/api/health | jq .

# Should respond with 200
# Confirms API is running and healthy
```

**Test 2: Check broadcast interval**
```bash
# Monitor WebSocket messages for 30 seconds
python3 << 'PYTHON'
import websocket
import json
import time

ws = websocket.create_connection("ws://localhost:8080/ws")
connected_users_msgs = []

start = time.time()
while time.time() - start < 30:
    try:
        msg = ws.recv()
        data = json.loads(msg)
        if data.get('type') == 'connected_users':
            connected_users_msgs.append(time.time())
    except:
        break
        
ws.close()

# Calculate intervals
if len(connected_users_msgs) > 1:
    intervals = [connected_users_msgs[i+1] - connected_users_msgs[i] 
                 for i in range(len(connected_users_msgs)-1)]
    avg_interval = sum(intervals) / len(intervals)
    print(f"Average interval: {avg_interval:.1f} seconds")
    print(f"Expected: 10 seconds")
else:
    print("Not enough messages to calculate interval")
PYTHON
```

**Expected Result:**
- Average interval ~10 seconds
- Variation of ±1-2 seconds is normal

### UI Testing with Gotenberg

**Test 1: Footer Display (Connected)**
```bash
# Navigate to page and take screenshot
# Footer should show user count
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000/ \
  --form width=1280 \
  --form height=768 \
  -o /tmp/footer-connected.png

# Verify screenshot contains "online"
file /tmp/footer-connected.png
```

**Verification:**
- [ ] Footer visible at bottom
- [ ] Green indicator present
- [ ] User count displayed
- [ ] "users online" text visible
- [ ] People icon rendered

**Test 2: Mobile View**
```bash
# Mobile width screenshot
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000/ \
  --form width=720 \
  --form height=1024 \
  -o /tmp/footer-mobile.png
```

**Verification:**
- [ ] Footer still visible on mobile
- [ ] Count still readable
- [ ] No overflow or wrapping issues
- [ ] Touch-friendly height

**Test 3: All Pages Show Footer**
```bash
# Check multiple pages
for page in / /countries /stats; do
  curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
    --form url=http://localhost:3000$page \
    --form width=1280 --form height=768 \
    -o /tmp/footer-$page.png
done

# All should have footer with user count
ls -l /tmp/footer-*.png
```

**Verification:**
- [ ] All pages have footer
- [ ] User count visible on each page
- [ ] Consistent styling across pages

### Integration Testing

**Full Integration Test:**
```bash
# 1. Restart all services
docker compose down
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d
sleep 5

# 2. Verify API running
curl -s http://localhost:8080/api/health | jq .

# 3. Test WebSocket
python3 << 'PYTHON'
import websocket
import json

ws = websocket.create_connection("ws://localhost:8080/ws")

# Receive first 3 messages
for i in range(3):
    msg = ws.recv()
    data = json.loads(msg)
    print(f"Message {i+1}: {data['type']}")
    
ws.close()
PYTHON

# 4. Take UI screenshot
curl --request POST http://localhost:3001/forms/chromium/screenshot/url \
  --form url=http://localhost:3000 \
  --form width=1280 --form height=768 \
  -o /tmp/full-integration.png

# 5. Check logs
docker compose logs leaderboard-api | grep -i "error\|ws" | head -10
```

## Database

### Data Source

Connected user count comes from:
- `websocket/hub.go` - In-memory client registry
- NOT from database (real-time, in-memory only)

### No Database Dependencies
- This feature is purely in-memory
- Tracks active connections during runtime only
- Not persisted to database

## WebSocket Hub Implementation

**Backend Details (for reference):**

### Hub Structure
```go
type Hub struct {
    clients    map[*Client]struct{}    // Connected clients
    broadcast  chan []byte             // Message broadcast channel
    register   chan *Client            // Register new client
    unregister chan *Client            // Unregister client
    mu         sync.RWMutex            // Thread-safe access
}
```

### ClientCount() Method
```go
func (h *Hub) ClientCount() int {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return len(h.clients)
}
```

This is called every 10 seconds to broadcast current count.

## Deployment Notes

### Prerequisites
- WebSocket support in API (gorilla/websocket package)
- CORS configured to accept WebSocket origins
- Port 8080 open for both HTTP and WebSocket traffic

### Build Process
```bash
cd /home/kumy/GIT/geokrety-points-system
docker compose build leaderboard-api leaderboard-dashboard
docker compose up -d
```

### Verification After Deploy
```bash
# 1. Check API health
curl -s http://localhost:8080/api/health | jq .

# 2. Test WebSocket
python3 << 'PYTHON'
import websocket
import json

try:
    ws = websocket.create_connection("ws://localhost:8080/ws", timeout=5)
    msg = ws.recv()
    data = json.loads(msg)
    print(f"✓ WebSocket working. Got: {data['type']}")
    ws.close()
except Exception as e:
    print(f"✗ WebSocket failed: {e}")
PYTHON

# 3. Check UI
curl -s http://localhost:3000 | grep -o "bi-people-fill" && echo "✓ Footer component loaded"
```

## Known Issues / Limitations

- **In-Memory Only** - Count resets on API restart (no persistence)
- **No User Identification** - Counts anonymous connections only
- **No Per-Route Tracking** - Count is global, not per page
- **Broadcast Frequency Fixed** - Cannot adjust from UI
- **No Maximum Limit** - Could theoretically accept unlimited connections
- **No Timeout Handling** - Very slow clients may not receive updates

## Future Enhancements

1. **Per-Page Tracking**
   - Show users on each specific page (/countries, /stats, etc.)
   - Track page-specific engagement

2. **User Identification**
   - Show number of authenticated vs anonymous users
   - Track specific user viewing patterns

3. **Connection Metrics**
   - Average session duration
   - Peak usage times
   - Geographic distribution (with privacy respect)

4. **Persistence**
   - Store connection history
   - Generate usage reports
   - Analyze engagement trends

5. **Alerts**
   - Notify admins of unusual activity
   - Connection spike warnings
   - Service health indicators

6. **Statistics Dashboard**
   - Dedicated metrics page
   - Connection history graphs
   - Performance metrics

## Related Features

- [country-leaderboard.md](country-leaderboard.md) - Country rankings
- [breakdown-charts.md](breakdown-charts.md) - Statistics visualizations

## Architectural Notes

**Message Flow:**
```
1. Client connects to /ws
2. Handler upgrades HTTP to WebSocket
3. Client registered in Hub.clients
4. StartBroadcaster running in background
   - Every 10 seconds:
     - Gets Hub.ClientCount()
     - Creates connected_users message
     - Broadcasts to all clients in Hub
5. Client receives message via WebSocket
6. Frontend updates reactive state
7. Vue template updates with new count
8. Footer re-renders with latest count
```

**Key Files:**
- `leaderboard-api/cmd/api/main.go` - Starts broadcaster
- `leaderboard-api/internal/handlers/websocket.go` - ServeWS handler, StartBroadcaster
- `leaderboard-api/internal/websocket/hub.go` - Hub implementation, ClientCount()
- `leaderboard-dashboard/src/composables/useWebSocket.js` - Frontend connection management
- `leaderboard-dashboard/src/App.vue` - Footer display

---

**Last Updated:** 2026-02-28  
**Version:** 1.0  
**Maintainer:** Development Team
