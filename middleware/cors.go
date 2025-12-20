package middleware

import (
	"Backend_Dorm_PTIT/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CORS(cfg *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := false
		for _, o := range cfg.AllowOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", strconv.FormatBool(cfg.AllowCreds))
			c.Header("Vary", "Origin")
		}

		// ⚠️ luôn set cho preflight
		c.Header("Access-Control-Allow-Headers", cfg.AllowHeaders)
		c.Header("Access-Control-Allow-Methods", cfg.AllowMethods)
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
