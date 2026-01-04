package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/constants"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/service"
	"Backend_Dorm_PTIT/utils"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	cfg          *config.Config
	wsSvc        *service.WSService
	contractRepo *repository.ContractRepository
}

func NewWSHandler(cfg *config.Config, wsSvc *service.WSService, contractRepo *repository.ContractRepository) *WSHandler {
	return &WSHandler{cfg: cfg, wsSvc: wsSvc, contractRepo: contractRepo}
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

// --- Chat over WebSocket ---

type ChatClientMessage struct {
	Type    string `json:"type"` // join_room, leave_room, chat_message
	Room    string `json:"room"`
	Content string `json:"content"`
}

// HandleWSChat quản lý kết nối chat (nhiều room trên 1 connection)
func (h *WSHandler) HandleWSChat(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error().Err(err).Msg("WebSocket upgrade failed (chat)")
		return
	}
	defer conn.Close()

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		logger.Error().Err(err).Msg("Unauthorized WebSocket chat connection attempt")
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized: "+err.Error()))
		return
	}

	conn.SetReadDeadline(time.Now().Add(time.Duration(h.cfg.WebSocket.PongWait) * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(time.Duration(h.cfg.WebSocket.PongWait) * time.Second))
		return nil
	})

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
					logger.Error().Err(err).Msg("Ping failed, closing chat connection")
					onceClose.Do(func() { close(done) })
					return
				}
			case <-done:
				return
			}
		}
	}()

	defer func() {
		h.wsSvc.LeaveAllRooms(conn)
		onceClose.Do(func() { close(done) })
	}()

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			logger.Info().Msgf("Chat WS user %s disconnected: %v", userID, err)
			return
		}
		if messageType != websocket.TextMessage {
			continue
		}

		// Heartbeat đơn giản
		if string(payload) == constants.HeartbeatCheck {
			conn.WriteMessage(websocket.TextMessage, []byte(constants.HeartbeatAck))
			continue
		}

		var msg ChatClientMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			logger.Error().Err(err).Msg("Invalid chat message JSON")
			conn.WriteJSON(map[string]string{
				"type":  "error",
				"error": "invalid_message",
			})
			continue
		}

		switch msg.Type {
		case "join_room":
			if msg.Room == "" {
				conn.WriteJSON(map[string]string{"type": "error", "error": "room_required"})
				continue
			}
			// Chỉ sinh viên có hợp đồng approved ở phòng đó mới được join
			ok, err := h.contractRepo.HasApprovedContractInRoom(c.Request.Context(), userID, msg.Room)
			if err != nil {
				logger.Error().Err(err).Msg("check HasApprovedContractInRoom failed")
				conn.WriteJSON(map[string]string{"type": "error", "error": "internal_error"})
				continue
			}
			if !ok {
				conn.WriteJSON(map[string]string{"type": "error", "error": "not_allowed"})
				continue
			}
			h.wsSvc.JoinRoom(msg.Room, userID, conn)
			conn.WriteJSON(map[string]string{"type": "joined", "room": msg.Room})

		case "leave_room":
			if msg.Room == "" {
				continue
			}
			h.wsSvc.LeaveRoom(msg.Room, conn)
			conn.WriteJSON(map[string]string{"type": "left", "room": msg.Room})

		case "chat_message":
			if msg.Room == "" || msg.Content == "" {
				continue
			}
			// Không cần check lại quyền nếu đã join room, nhưng có thể check bổ sung nếu muốn
			h.wsSvc.BroadcastToRoom(msg.Room, userID, msg.Content)

		default:
			conn.WriteJSON(map[string]string{"type": "error", "error": "unknown_type"})
		}
	}
}
