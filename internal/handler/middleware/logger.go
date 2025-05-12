package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		mwLog := log.WithGroup("HTTP")

		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set(RequestIDKey, reqID)
		c.Writer.Header().Set("X-Request-ID", reqID)

		baseLog := mwLog.With(
			slog.String("path", c.Request.URL.RequestURI()),
			slog.String("method", c.Request.Method),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("request_id", reqID),
		)

		start := time.Now()
		startLog := baseLog.With(
			slog.Time("start_time", start),
		)

		startLog.InfoContext(c, "Request started")
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		completeLog := baseLog.With(
			slog.Int("status_code", statusCode),
			slog.Duration("latency_time", latency),
		)

		errors := c.Errors.ByType(gin.ErrorTypeAny).Errors()
		if len(errors) > 0 {
			msgs := strings.Join(errors, "; ")
			completeLog = completeLog.With(slog.String("error_messages", msgs))
		}

		if statusCode >= 500 {
			completeLog.ErrorContext(c, "Request failed")
		} else if statusCode >= 400 {
			completeLog.WarnContext(c, "Request failed")
		} else {
			completeLog.InfoContext(c, "Request completed")
		}
	}
}
