package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/constants"
	"Backend_Dorm_PTIT/database"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"fmt"

	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
}

// MicrosoftOAuthRequest is the body sent from FE with Microsoft access_token
type MicrosoftOAuthRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

func NewAuthHandler(cfg *config.Config, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{cfg: cfg, userRepo: userRepo}
}

// MicrosoftOAuthCallback handles Microsoft OAuth2 login via access_token from FE
// FE gọi Microsoft trực tiếp để lấy access_token, sau đó gửi access_token cho BE
// BE dùng access_token gọi Graph lấy profile, rồi login theo email
func (h *AuthHandler) MicrosoftOAuthCallback(c *gin.Context) {
	var req MicrosoftOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "invalid request: "+err.Error()))
		return
	}

	ctx := c.Request.Context()
	msAccessToken := req.AccessToken

	// Call Microsoft Graph to get user profile
	client := &http.Client{Timeout: 10 * time.Second}
	profileReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to create profile request: "+err.Error()))
		return
	}
	profileReq.Header.Set("Authorization", "Bearer "+msAccessToken)
	respProfile, err := client.Do(profileReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse(http.StatusBadGateway, "failed to call Microsoft Graph: "+err.Error()))
		return
	}
	defer respProfile.Body.Close()

	if respProfile.StatusCode < 200 || respProfile.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(respProfile.Body)
		c.JSON(http.StatusBadGateway, models.ErrorResponse(http.StatusBadGateway, fmt.Sprintf("graph api error: %d - %s", respProfile.StatusCode, string(bodyBytes))))
		return
	}

	var user map[string]interface{}
	if err := json.NewDecoder(respProfile.Body).Decode(&user); err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse(http.StatusBadGateway, "failed to decode graph user: "+err.Error()))
		return
	}

	displayName, _ := user["displayName"].(string)
	email := ""
	if m, ok := user["mail"].(string); ok && m != "" {
		email = m
	} else if upn, ok := user["userPrincipalName"].(string); ok {
		email = upn
	}
	if email == "" {
		c.JSON(http.StatusBadGateway, models.ErrorResponse(http.StatusBadGateway, "could not determine email from microsoft profile"))
		return
	}

	// Chỉ chấp nhận email sinh viên PTIT
	lowerEmail := strings.ToLower(email)
	if !strings.HasSuffix(lowerEmail, "@stu.ptit.edu.vn") {
		c.JSON(http.StatusForbidden, models.ErrorResponse(http.StatusForbidden, "email must be a PTIT student email (@stu.ptit.edu.vn)"))
		return
	}

	// Từ đây coi như mật khẩu đã hợp lệ, tiến hành login theo email
	config := h.cfg
	jSecret := config.JWT.Secret
	if jSecret == "" {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "jwt secret not configured"))
		return
	}

	// Tìm user theo email
	userRecord, err := h.userRepo.GetByEmail(ctx, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to get user by email: "+err.Error()))
		return
	}

	var userInfo models.LoginUserInfo
	// Helper check role
	hasRole := func(roles []string, name string) bool {
		for _, r := range roles {
			if r == name {
				return true
			}
		}
		return false
	}

	if userRecord == nil {
		// Chưa tồn tại user -> tạo user mới role guest
		newID := uuid.NewString()
		username := strings.SplitN(lowerEmail, "@", 2)[0]
		now := time.Now()
		newUser := &models.User{
			ID:           newID,
			Email:        email,
			Username:     username,
			PasswordHash: "", // không dùng password cho OAuth
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := h.userRepo.Create(ctx, newUser); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to create user: "+err.Error()))
			return
		}
		if err := h.userRepo.SetUserRoleByName(ctx, newID, "guest"); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to assign guest role: "+err.Error()))
			return
		}
		userInfo = models.LoginUserInfo{
			UserID:      newID,
			Email:       email,
			Username:    username,
			DisplayName: displayName,
			Avatar:      "",
			Roles:       []string{"guest"},
		}
	} else {
		// User đã tồn tại: quyết định role theo DB
		roles, err := h.userRepo.GetRolesByUserID(ctx, userRecord.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to get user roles: "+err.Error()))
			return
		}

		// Ưu tiên student, sau đó guest; các role khác không cho đăng nhập bằng Microsoft
		var loginRoles []string
		if hasRole(roles, "student") {
			loginRoles = []string{"student"}
		} else if hasRole(roles, "guest") || len(roles) == 0 {
			loginRoles = []string{"guest"}
		} else {
			c.JSON(http.StatusForbidden, models.ErrorResponse(http.StatusForbidden, "account role is not allowed for Microsoft login"))
			return
		}

		// Thử lấy thông tin hiển thị từ hệ thống nếu có, nếu lỗi thì fallback
		if uInfo, err := h.userRepo.GetUserInfo(ctx, userRecord.Username); err == nil {
			userInfo = models.LoginUserInfo{
				UserID:      uInfo.UserID,
				Email:       uInfo.Email,
				Username:    uInfo.Username,
				DisplayName: uInfo.DisplayName,
				Avatar:      uInfo.Avatar,
				Roles:       loginRoles,
			}
		} else {
			userInfo = models.LoginUserInfo{
				UserID:      userRecord.ID,
				Email:       userRecord.Email,
				Username:    userRecord.Username,
				DisplayName: displayName,
				Avatar:      "",
				Roles:       loginRoles,
			}
		}
	}

	// Sinh cặp access/refresh token giống luồng login thường
	tokenID := uuid.NewString()
	accessClaims := jwt.MapClaims{
		"token_id": tokenID,
		"user_id":  userInfo.UserID,
		"roles":    userInfo.Roles,
		"type":     "access",
		"exp":      time.Now().Add(time.Duration(config.JWT.Access_Exp) * time.Second).Unix(),
	}

	accessTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessTokenJWT.SignedString([]byte(jSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to sign access token: "+err.Error()))
		return
	}

	refreshClaims := jwt.MapClaims{
		"token_id": tokenID,
		"user_id":  userInfo.UserID,
		"roles":    userInfo.Roles,
		"type":     "refresh",
		"exp":      time.Now().Add(time.Duration(config.JWT.Refresh_Exp) * time.Second).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString([]byte(jSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to sign refresh token: "+err.Error()))
		return
	}

	tokenTTL := time.Duration(config.JWT.Refresh_Exp) * time.Second
	if err := database.Set(tokenID, userInfo.UserID, tokenTTL); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to store token: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		User:         userInfo,
	})
}

// LogoutHandler godoc
// @Summary Logout and revoke refresh token
// @Description Invalidate the provided refresh token and remove it from Redis whitelist. Always returns HTTP 200 with different message responses:
// @Tags Auth
// @Accept json
// @Produce json
// @Param LogoutRequest body models.LogoutRequest true "Refresh token to revoke"
// @Success 200 {object} models.Response "Logout status message (see Description for possible values)"
// @Router /logout [post]
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	logger.Info().Msg("Logout request received")

	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to bind logout request")
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidMissingRefreshToken,
			Data:    nil,
		})
		return
	}

	secret := h.cfg.JWT.Secret
	if secret == "" {
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidJWTSecretNotConfigured,
			Data:    nil,
		})
		return
	}

	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidInvalidRefreshToken,
			Data:    nil,
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidInvalidClaims,
			Data:    nil,
		})
		return
	}

	if t, _ := claims["type"].(string); t != "refresh" {
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidNotRefreshToken,
			Data:    nil,
		})
		return
	}

	oldTokenID, _ := claims["token_id"].(string)

	exists, _, err := database.Get(oldTokenID)
	if err != nil || !exists {
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidTokenIDNotFound,
			Data:    nil,
		})
		return
	}

	if err := database.Delete(oldTokenID); err != nil {
		logger.Error().Err(err).Str("token_id", oldTokenID).Msg("Failed to delete token from Redis")
		c.JSON(http.StatusOK, models.Response{
			Code:    http.StatusOK,
			Message: constants.MsgLogoutSuccessButTokenInvalidFailedToDelete,
			Data:    nil,
		})
		return
	}

	logger.Info().Str("token_id", oldTokenID).Msg("User logged out successfully")
	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: constants.MsgLogoutSuccessTokenDeleted,
		Data:    nil,
	})
}

// LogoutAllSessionsHandler godoc
// @Summary Logout all sessions for a user
// @Description Xác thực username/password, sau đó logout toàn bộ session (xóa hết token của user)
// @Tags Auth
// @Accept json
// @Produce json
// @Param LogoutAllSessionsRequest body LogoutAllSessionsRequest true "Username và password để xác thực"
// @Success 200 {object} models.Response "Logout all sessions success"
// @Failure 400 {object} models.Response "Missing or invalid request"
// @Failure 401 {object} models.Response "Invalid username or password"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /logout-all [post]
func (h *AuthHandler) LogoutAllSessionsHandler(c *gin.Context) {
	var req models.LogoutAllSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "missing username or password"))
		return
	}
	ok, err := h.userRepo.VerifyCredentials(c.Request.Context(), req.Username, req.Password)
	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "invalid username or password"))
		return
	}
	user, err := h.userRepo.GetUserInfo(c.Request.Context(), req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "user not found"))
		return
	}
	err = database.DeleteAllTokensByUserID(user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to logout all sessions"))
		return
	}
	c.JSON(http.StatusOK, models.Response{
		Code:    http.StatusOK,
		Message: "All sessions logged out successfully",
		Data:    nil,
	})
}

// LoginHandler godoc
// @Summary Login login
// @Description Exchange Login code for access/refresh tokens, fetch user info, and store token in Redis whitelist. Returns JWT tokens and user info.
// @Tags Auth
// @Accept json
// @Produce json
// @Param LoginRequest body models.LoginRequest true "Login code and redirect URI"
// @Success 200 {object} models.LoginResponse "Access and Refresh token (JWT) and user info"
// @Failure 400 {object} models.Response "Missing code or invalid request"
// @Failure 502 {object} models.Response "Failed to exchange code or get user info"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /Login [post]
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "missing username or password"))
		return
	}

	// Kiểm tra username, password qua repo
	ok, err := h.userRepo.VerifyCredentials(c.Request.Context(), req.Username, req.Password)
	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "invalid username or password"))
		return
	}

	// Kiểm tra trạng thái toàn khoản nếu là inactive thì k cho đăng nhập , lấy status qua username từ repo
	status, err := h.userRepo.GetStatusByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to get user status"))
		return
	}
	if status == "inactive" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "user account is inactive"))
		return
	}

	// Lấy user info qua email
	user, err := h.userRepo.GetUserInfo(c.Request.Context(), req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "user not found"))
		return
	}

	userInfo := models.LoginUserInfo{
		UserID:      user.UserID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		Roles:       user.Roles,
	}

	config := h.cfg
	jwtSecret := config.JWT.Secret
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "jwt secret not configured"))
		return
	}

	tokenID := uuid.NewString()

	accessClaims := jwt.MapClaims{
		"token_id": tokenID,
		"user_id":  userInfo.UserID,
		"roles":    userInfo.Roles,
		"type":     "access",
		"exp":      time.Now().Add(time.Duration(config.JWT.Access_Exp) * time.Second).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to sign access token: "+err.Error()))
		return
	}

	refreshClaims := jwt.MapClaims{
		"token_id": tokenID,
		"user_id":  userInfo.UserID,
		"roles":    userInfo.Roles,
		"type":     "refresh",
		"exp":      time.Now().Add(time.Duration(config.JWT.Refresh_Exp) * time.Second).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to sign refresh token: "+err.Error()))
		return
	}

	tokenTTL := time.Duration(config.JWT.Refresh_Exp) * time.Second
	if err := database.Set(tokenID, userInfo.UserID, tokenTTL); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to store token: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		User:         userInfo,
	})
}

// RefreshHandler godoc
// @Summary Refresh access token
// @Description Validate refresh token, rotate token ID, issue new access/refresh tokens, and update Redis whitelist. Returns new JWT tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param RefreshRequest body models.RefreshRequest true "Refresh token to validate"
// @Success 200 {object} models.RefreshResponse "New JWT tokens"
// @Failure 400 {object} models.Response "Missing or invalid refresh token"
// @Failure 401 {object} models.Response "Unauthorized or token not found"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /refresh [post]
func (h *AuthHandler) RefreshHandler(c *gin.Context) {
	logger.Info().Msg("Token refresh request received")

	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to bind refresh request")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "missing refresh_token"))
		return
	}

	reqData, _ := json.Marshal(req)
	hashRequest := fmt.Sprintf("refresh_req:%x", sha256.Sum256(reqData))
	hashLockKey := fmt.Sprintf("refresh_lock:%x", sha256.Sum256(reqData))

	lockKey := hashLockKey
	maxRetry := 25
	lockAcquired := false
	for i := 0; i < maxRetry; i++ {
		ok, err := database.SetLockKey(lockKey, "1", 25*time.Second)
		if err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Lock Key is being held!")
		}
		if ok {
			lockAcquired = true
			logger.Info().Str("lockKey", lockKey).Msg("Lock acquired")
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
	if !lockAcquired {
		logger.Error().Str("lockKey", lockKey).Msg("Could not acquire lock after retries")
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse(http.StatusTooManyRequests, "Server busy, please retry"))
		return
	}

	if ok, cachedResp, err := database.GetCacheRequest(hashRequest); err == nil && ok {
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		logger.Info().Str("hash_request", hashRequest).Msg("Cache request exists, returning cached response")
		c.Data(http.StatusOK, "application/json", []byte(cachedResp))
		return
	}

	config := h.cfg
	secret := config.JWT.Secret
	if secret == "" {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "jwt secret not configured"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	token, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, constants.ErrInvalidOrExpiredRefreshToken))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "invalid claims"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	if t, _ := claims["type"].(string); t != "refresh" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "token is not a refresh token"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	userID, _ := claims["user_id"].(string)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "invalid refresh token (missing user_id)"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}
	rolesRaw, ok := claims["roles"]
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "invalid refresh token (missing roles)"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}
	var Roles []string
	switch v := rolesRaw.(type) {
	case []interface{}:
		for _, r := range v {
			if s, ok := r.(string); ok {
				Roles = append(Roles, s)
			}
		}
	case []string:
		Roles = v
	}
	if len(Roles) == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "invalid refresh token (missing roles)"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	oldTokenID, _ := claims["token_id"].(string)

	exists, _, err := database.Get(oldTokenID)
	if err != nil || !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "refresh token not found on whitelist"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	if err := database.Delete(oldTokenID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to delete old refresh token"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	newTokenID := uuid.NewString()
	accessClaims := jwt.MapClaims{
		"token_id": newTokenID,
		"user_id":  userID,
		"roles":    Roles,
		"type":     "access",
		"exp":      time.Now().Add(time.Duration(config.JWT.Access_Exp) * time.Second).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to sign access token"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	refreshClaims := jwt.MapClaims{
		"token_id": newTokenID,
		"user_id":  userID,
		"roles":    Roles,
		"type":     "refresh",
		"exp":      time.Now().Add(time.Duration(config.JWT.Refresh_Exp) * time.Second).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to sign refresh token"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}

	tokenTTL := time.Duration(config.JWT.Refresh_Exp) * time.Second

	if err := database.Set(newTokenID, userID, tokenTTL); err != nil {
		logger.Error().Err(err).Str("token_id", newTokenID).Str("user_id", userID).Msg("Failed to store new refresh token")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "failed to store new refresh token"))
		if err := database.DeleteLockKey(lockKey); err != nil {
			logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
		} else {
			logger.Info().Str("lockKey", lockKey).Msg("Lock released")
		}
		return
	}
	resp := models.RefreshResponse{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		UserID:       userID,
	}

	respData, _ := json.Marshal(resp)
	if err := database.SetCacheRequest(hashRequest, string(respData), 30*time.Second); err != nil {
		logger.Error().Err(err).Str("hash_request", hashRequest).Msg("Failed to cache refresh response")
	} else {
		logger.Info().Str("hash_request", hashRequest).Msg("Refresh response cache saved successfully")
	}

	if err := database.DeleteLockKey(lockKey); err != nil {
		logger.Error().Err(err).Str("lockKey", lockKey).Msg("Failed to release lock")
	} else {
		logger.Info().Str("lockKey", lockKey).Msg("Lock released")
	}

	logger.Info().
		Str("user_id", userID).
		Str("old_token_id", oldTokenID).
		Str("new_token_id", newTokenID).
		Msg("Token refreshed successfully")

	c.JSON(http.StatusOK, resp)
}
