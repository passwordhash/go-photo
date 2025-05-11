package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		start := time.Now()
		userAgent := c.Request.UserAgent()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path += "?" + raw
		}

		startFields := log.Fields{
			"path":       path,
			"method":     c.Request.Method,
			"client_ip":  clientIP,
			"start_time": start.Format(time.DateTime),
			"user_agent": userAgent,
		}

		log.WithFields(startFields).Info("Request started")
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errorMessages := c.Errors.ByType(gin.ErrorTypePrivate).Errors()
		errorMessage := strings.Join(errorMessages, "; ")

		completeFields := log.Fields{
			"status_code":  statusCode,
			"path":         path,
			"method":       c.Request.Method,
			"client_ip":    clientIP,
			"latency_time": latency,
			"user_agent":   userAgent,
		}

		entry := log.WithFields(completeFields)
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
