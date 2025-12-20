package middleware

import (
	"net/http"
	"strings"
    "Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/constants"
	"Backend_Dorm_PTIT/database"
	"Backend_Dorm_PTIT/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Authentication returns a middleware that validates JWT tokens and checks whitelist
func Authentication(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			abortWithError(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}

		if jwtSecret == "" {
			abortWithError(c, http.StatusInternalServerError, "JWT secret not configured")
			return
		}

		token, err := parseToken(tokenString, jwtSecret)
		if err != nil {
			abortWithError(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			abortWithError(c, http.StatusUnauthorized, constants.ErrInvalidToken)
			return
		}

		if err := validateClaims(claims); err != nil {
			abortWithError(c, http.StatusUnauthorized, err.Error())
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}

func AuthenticateWS(jwtSecret string) gin.HandlerFunc {
       return func(c *gin.Context) {
	       tokenString := c.Query("token")
	       if tokenString == "" {
			   logger.Error().Msg("Missing token in query param")
		       c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			   return 
	       }
	       if jwtSecret == "" {
			   	logger.Error().Msg("JWT secret not configured")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
				return
			}
	       token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		       if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			       return nil, jwt.ErrSignatureInvalid
		       }
		       return []byte(jwtSecret), nil
	       })
	       if err != nil {
			   logger.Error().Err(err).Msg("Invalid token")	
		       c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			   return
	       }
	       claims, ok := token.Claims.(jwt.MapClaims)
	       if !ok || !token.Valid {
			   logger.Error().Msg("Invalid claims in token")
		       c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			   return
	       }

	       tokenType, _ := claims["type"].(string)
	       if tokenType != "access" {
			   logger.Error().Msg("Token is not an access token")
		       c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is not an access token"})
		       return
	       }
	       tokenID, _ := claims["token_id"].(string)
	       if tokenID == "" {
			   logger.Error().Msg("Token is missing token_id")
		       c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is missing token_id"})
		       return
	       }
		   ok, _, err = database.Get(tokenID)
	       if err != nil {
			   logger.Error().Err(err).Msg("Error checking whitelist")
		       c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error checking whitelist"})
		       return
	       }
	       if !ok {
			   logger.Error().Msg("Token is not in whitelist")
		       c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is not in whitelist"})
		       return
	       }
	       c.Set("user", claims)
	       c.Next()
       }
}

// extractBearerToken extracts the token from the Authorization header
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	token, found := strings.CutPrefix(authHeader, "Bearer ")
	if found {
		return token
	}
	return ""
}

// parseToken parses and validates the JWT token
func parseToken(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
}

// validateClaims validates the token claims
func validateClaims(claims jwt.MapClaims) error {
	// Check token type
	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return jwt.ErrTokenInvalidClaims
	}

	// Check token ID
	tokenID, _ := claims["token_id"].(string)
	if tokenID == "" {
		return jwt.ErrTokenInvalidClaims
	}

	// Check whitelist
	ok, _, err := database.Get(tokenID)
	if err != nil || !ok {
		return jwt.ErrTokenInvalidClaims
	}

	return nil
}

// abortWithError aborts the request with a standardized error response
func abortWithError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, models.ErrorResponse(statusCode, message))
}