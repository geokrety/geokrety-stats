package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	wsHub "github.com/geokrety/leaderboard-api/internal/websocket"
	"github.com/geokrety/leaderboard-api/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled by middleware
	},
}

// ServeWS handles GET /ws – upgrades the connection and registers it with the hub.
func (h *Handler) ServeWS(hub *wsHub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error().Err(err).Msg("ws upgrade failed")
			return
		}

		// Push current snapshot immediately upon connection
		go func() {
			time.Sleep(100 * time.Millisecond)
			if entries, err := h.TopLeaderboard(context.Background(), 10); err == nil {
				hub.Broadcast(models.WSMessage{Type: "leaderboard_snapshot", Payload: entries})
			}
			stats := h.computeGlobalStatsFallback(context.Background())
			hub.Broadcast(models.WSMessage{Type: "global_stats", Payload: stats})
		}()

		hub.ServeClient(conn)
	}
}

// StartBroadcaster periodically pushes leaderboard updates to all connected clients.
func (h *Handler) StartBroadcaster(ctx context.Context, hub *wsHub.Hub, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if hub.ClientCount() == 0 {
				continue
			}
			entries, err := h.TopLeaderboard(ctx, 10)
			if err != nil {
				log.Error().Err(err).Msg("ws: leaderboard broadcast failed")
				continue
			}
			hub.Broadcast(models.WSMessage{
				Type:    "leaderboard_update",
				Payload: entries,
			})
		}
	}
}
