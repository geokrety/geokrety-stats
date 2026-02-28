import { ref, onMounted, onUnmounted } from 'vue'

const WS_URL = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`

// Global reactive state shared across components
const connected = ref(false)
const leaderboard = ref([])
const stats = ref(null)
const connectedUsers = ref(0)

let ws = null
let reconnectTimer = null

function connect() {
  if (ws && ws.readyState < 2) return

  ws = new WebSocket(WS_URL)

  ws.onopen = () => {
    connected.value = true
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  }

  ws.onclose = () => {
    connected.value = false
    reconnectTimer = setTimeout(connect, 5000)
  }

  ws.onerror = () => {
    ws.close()
  }

  ws.onmessage = (evt) => {
    try {
      const msg = JSON.parse(evt.data)
      if (msg.type === 'leaderboard_update' || msg.type === 'leaderboard_snapshot') {
        leaderboard.value = msg.payload ?? []
      } else if (msg.type === 'global_stats') {
        stats.value = msg.payload
      } else if (msg.type === 'connected_users') {
        connectedUsers.value = msg.payload?.count ?? 0
      }
    } catch (_) { /* noop */ }
  }
}

function disconnect() {
  if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  if (ws) { ws.onclose = null; ws.close(); ws = null }
  connected.value = false
}

/**
 * Composable — call once at app root (App.vue) to start the WS connection.
 * Subsequent calls from other components just read the reactive state.
 */
export function useLiveStats() {
  onMounted(() => {
    if (!ws) connect()
  })
  onUnmounted(() => {
    // only disconnect if this is the last consumer (App.vue)
  })
  return { connected, leaderboard, stats, connectedUsers }
}

/**
 * Use leaderboard snapshot without managing connection lifecycle.
 */
export function useLeaderboardLive() {
  return { connected, leaderboard }
}
