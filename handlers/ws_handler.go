package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/constants"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/service"
	"Backend_Dorm_PTIT/utils"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	cfg   *config.Config
	wsSvc *service.WSService
}

func NewWSHandler(cfg *config.Config, wsSvc *service.WSService) *WSHandler {
	return &WSHandler{cfg: cfg, wsSvc: wsSvc}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WSHandler) HandleWSAdmin(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}
	defer conn.Close()
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("Unauthorized WebSocket connection attempt")
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized: "+err.Error()))
		return
	}

	conn.SetReadDeadline(time.Now().Add(time.Duration(h.cfg.WebSocket.PongWait) * time.Second))
	conn.SetPongHandler(func(string) error {
		logger.Info().Msgf("Pong received from user %s", userID)
		conn.SetReadDeadline(time.Now().Add(time.Duration(h.cfg.WebSocket.PongWait) * time.Second))
		return nil
	})

	h.wsSvc.AddConnection(userID, conn)
	defer func() {
		h.wsSvc.RemoveConnection(userID, conn)
	}()

	done := make(chan struct{})
	var onceClose sync.Once

	// Goroutine: gửi ping định kỳ
	go func() {
		ticker := time.NewTicker(time.Duration(h.cfg.WebSocket.PingPeriod) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(time.Duration(h.cfg.WebSocket.WriteWait) * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					logger.Error().Err(err).Msg("Ping failed, closing connection")
					onceClose.Do(func() { close(done) })
					return
				} else {
					logger.Info().Msgf("Ping sent to user %s", userID)
				}
			case <-done:
				return
			}
		}
	}()

	// Goroutine: stream log hệ thống về client, dừng bằng context
	logPath := h.cfg.Logging.FilePath
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	go service.StartTailLogFile(ctx, logPath, func(line string) {
		conn.SetWriteDeadline(time.Now().Add(time.Duration(h.cfg.WebSocket.WriteWait) * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			logger.Error().Err(err).Msg("Send log to ws client failed")
			onceClose.Do(func() { cancel(); close(done) })
		}
	})

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logger.Info().Msgf("User %s disconnected", userID)
			onceClose.Do(func() { cancel(); close(done) })
			return
		}
		if messageType == websocket.TextMessage && string(message) == constants.HeartbeatCheck {
			conn.WriteMessage(websocket.TextMessage, []byte(constants.HeartbeatAck))
			continue
		}
	}
}
