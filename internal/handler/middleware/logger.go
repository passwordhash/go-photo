package middleware

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path += "?" + raw
		}

		fields := log.Fields{
			"path":       path,
			"method":     c.Request.Method,
			"client_ip":  clientIP,
			"start_time": start.Format(time.DateTime),
		}

		log.WithFields(fields).Info("Request started")
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		fields = log.Fields{
			"status_code":  statusCode,
			"path":         path,
			"method":       c.Request.Method,
			"client_ip":    clientIP,
			"latency_time": latency,
		}

		entry := log.WithFields(fields)
		if errorMessage != "" {
			if statusCode >= 500 {
				entry.Error(errorMessage)
			} else if statusCode >= 400 {
				entry.Warn(errorMessage)
			}
		} else {
			entry.Info("Request completed")
		}
	}
}
