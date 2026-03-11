package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/geokrety/geokrety-stats-api/internal/metrics"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// TODO "nhooyr.io/websocket" is deprecated, consider switching to "https://github.com/coder/websocket"

type Envelope struct {
	Type string      `json:"type"`
	Path string      `json:"path"`
	Data interface{} `json:"data"`
}

type Hub struct {
	logger            *zap.Logger
	metrics           *metrics.Collector
	broadcastInterval time.Duration

	mu          sync.RWMutex
	connections map[*websocket.Conn]struct{}
	count       atomic.Int64
}

func NewHub(logger *zap.Logger, metricsCollector *metrics.Collector, broadcastInterval time.Duration) *Hub {
	return &Hub{
		logger:            logger,
		metrics:           metricsCollector,
		broadcastInterval: broadcastInterval,
		connections:       make(map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		h.logger.Warn("failed to accept websocket", zap.Error(err))
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "connection closed")

	h.addConn(conn)
	defer h.removeConn(conn)

	h.broadcastConnCount(r.Context())
	_ = h.broadcastInterval

	for {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		_, _, readErr := conn.Read(ctx)
		cancel()
		if readErr != nil {
			h.logger.Debug("websocket read loop ended", zap.Error(readErr))
			break
		}
	}
}

func (h *Hub) Publish(ctx context.Context, envelope Envelope) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.connections))
	for conn := range h.connections {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := wsjson.Write(writeCtx, conn, envelope)
		cancel()
		if err != nil {
			h.logger.Debug("failed websocket write", zap.Error(err))
			continue
		}
		h.metrics.WSBroadcastsTotal.Inc()
	}
}

func (h *Hub) PublishRaw(ctx context.Context, msgType, path string, payload json.RawMessage) {
	data := map[string]interface{}{}
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &data)
	}
	h.Publish(ctx, Envelope{Type: msgType, Path: path, Data: data})
}

func (h *Hub) ActiveConnections() int64 {
	return h.count.Load()
}

func (h *Hub) addConn(conn *websocket.Conn) {
	h.mu.Lock()
	h.connections[conn] = struct{}{}
	h.mu.Unlock()

	newCount := h.count.Add(1)
	h.metrics.WSConnections.Set(float64(newCount))
	h.logger.Info("websocket connected", zap.Int64("active_connections", newCount))
}

func (h *Hub) removeConn(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.connections, conn)
	h.mu.Unlock()

	newCount := h.count.Add(-1)
	if newCount < 0 {
		h.count.Store(0)
		newCount = 0
	}
	h.metrics.WSConnections.Set(float64(newCount))
	h.logger.Info("websocket disconnected", zap.Int64("active_connections", newCount))
	h.broadcastConnCount(context.Background())
}

func (h *Hub) broadcastConnCount(ctx context.Context) {
	h.Publish(ctx, Envelope{
		Type: "conn_count",
		Path: "global",
		Data: map[string]interface{}{
			"activeConnections": h.ActiveConnections(),
		},
	})
}
