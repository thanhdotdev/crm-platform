package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/bship1/bship-common-go.git/pkg/zlog"
)

// Logger returns a middleware that logs HTTP requests.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		zlog.Info("HTTP Request",
			zlog.Field("method", method),
			zlog.Field("path", path),
			zlog.Field("status", status),
			zlog.Field("latency", latency.String()),
		)
	}
}
