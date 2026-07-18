package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		allowOrigin := allowedOrigin(origin, allowedOrigins)
		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With,x-cookie-auth,xxx-api-auth,xxx_api_auth")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func allowedOrigin(origin string, allowedOrigins []string) string {
	if origin == "" {
		return ""
	}
	if len(allowedOrigins) == 0 {
		return origin
	}
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		if allowed == "*" || strings.EqualFold(allowed, origin) {
			return origin
		}
	}
	return ""
}
