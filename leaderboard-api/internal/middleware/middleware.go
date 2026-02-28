package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CORS returns a simple CORS middleware that accepts the provided origins.
func CORS(allowOrigins []string) gin.HandlerFunc {
	origins := make(map[string]struct{}, len(allowOrigins))
	for _, o := range allowOrigins {
		origins[o] = struct{}{}
	}
	allowAll := false
	if _, ok := origins["*"]; ok {
		allowAll = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := ""
		if allowAll {
			allowed = "*"
		} else if _, ok := origins[origin]; ok {
			allowed = origin
		}

		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Logger returns a structured request logger middleware.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Dur("latency", time.Since(start)).
			Msg("request")
	}
}
