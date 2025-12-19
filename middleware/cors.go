package middleware

import (
	"Backend_Dorm_PTIT/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CORS(corsConfig *config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := false
		for _, o := range corsConfig.AllowOrigins {
			if o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set(
				"Access-Control-Allow-Credentials",
				strconv.FormatBool(corsConfig.AllowCreds),
			)
			c.Writer.Header().Set(
				"Access-Control-Allow-Headers",
				corsConfig.AllowHeaders,
			)
			c.Writer.Header().Set(
				"Access-Control-Allow-Methods",
				corsConfig.AllowMethods,
			)
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
