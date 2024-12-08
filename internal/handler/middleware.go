package handler

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		entry := log.WithFields(log.Fields{
			"method":       method,
			"path":         path,
			"status_code":  statusCode,
			"client_ip":    clientIP,
			"latency_time": latency,
		})

		if errorMessage != "" {
			entry.Error(errorMessage)
		} else {
			entry.Info("Request processed")
		}
	}
}
