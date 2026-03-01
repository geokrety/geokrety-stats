import { ref, onMounted, onUnmounted } from 'vue'

const WS_URL = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`

// Global reactive state shared across components
const connected = ref(false)
const enabled = ref(localStorage.getItem('ws_enabled') !== 'false')
const leaderboard = ref([])
const stats = ref(null)
const connectedUsers = ref(0)
const lastUpdate = ref(0)

let ws = null
let reconnectTimer = null

function connect() {
  if (!enabled.value) return
  if (ws && ws.readyState < 2) return

  ws = new WebSocket(WS_URL)

  ws.onopen = () => {
    connected.value = true
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
    console.log('[WS] Connected to Geokrety WebSocket')
  }

  ws.onclose = () => {
    connected.value = false
    if (enabled.value) {
      reconnectTimer = setTimeout(connect, 5000)
    }
    console.log('[WS] Disconnected from Geokrety WebSocket')
  }

  ws.onerror = (err) => {
    console.error('[WS] WebSocket error:', err)
    ws.close()
  }

  ws.onmessage = (evt) => {
    try {
      const msg = JSON.parse(evt.data)
      console.log(`[WS] Received message: ${msg.type}`, msg.payload)
      if (msg.type === 'leaderboard_update' || msg.type === 'leaderboard_snapshot') {
        leaderboard.value = msg.payload ?? []
      } else if (msg.type === 'global_stats') {
        stats.value = msg.payload
        lastUpdate.value = Date.now()
      } else if (msg.type === 'connected_users') {
        connectedUsers.value = msg.payload?.count ?? 0
      }
    } catch (e) {
      console.error('[WS] Failed to parse message:', e)
    }
  }
}

function disconnect() {
  if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  if (ws) {
    ws.onclose = null
    ws.close()
    ws = null
  }
  connected.value = false
}

function toggleEnabled() {
  enabled.value = !enabled.value
  localStorage.setItem('ws_enabled', enabled.value)
  if (enabled.value) {
    connect()
  } else {
    disconnect()
  }
}

/**
 * Composable — call once at app root (App.vue) to start the WS connection.
 * Subsequent calls from other components just read the reactive state.
 */
export function useLiveStats() {
  onMounted(() => {
    if (!ws && enabled.value) connect()
  })
  return { connected, enabled, stats, connectedUsers, lastUpdate, toggleEnabled }
}

/**
 * Use leaderboard snapshot without managing connection lifecycle.
 */
export function useLeaderboardLive() {
  return { connected, leaderboard }
}
