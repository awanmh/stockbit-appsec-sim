package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.String("latency", latency.String()),
			slog.String("client_ip", clientIP),
		)
	}
}
