package service

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSService struct {
	mu sync.RWMutex
	// connections dùng cho luồng WS admin (log)
	connections map[string][]*websocket.Conn
	// rooms: roomID -> (conn -> userID)
	rooms map[string]map[*websocket.Conn]string
	// connRooms: conn -> set(roomID)
	connRooms map[*websocket.Conn]map[string]struct{}
	// TODO a Hoàng: có thể dùng map.sync sau.
}

func NewWSService() *WSService {
	return &WSService{
		connections: make(map[string][]*websocket.Conn),
		rooms:       make(map[string]map[*websocket.Conn]string),
		connRooms:   make(map[*websocket.Conn]map[string]struct{}),
	}
}

func (s *WSService) AddConnection(userID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[userID] = append(s.connections[userID], conn)
}

func (s *WSService) RemoveConnection(userID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conns := s.connections[userID]
	newConns := make([]*websocket.Conn, 0, len(conns))
	for _, c := range conns {
		if c != conn {
			newConns = append(newConns, c)
		} else {
			c.Close()
		}
	}
	if len(newConns) == 0 {
		delete(s.connections, userID)
	} else {
		s.connections[userID] = newConns
	}
}

func (s *WSService) GetConnections(userID string) ([]*websocket.Conn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conns, ok := s.connections[userID]
	return conns, ok
}

// --- Chat rooms helpers ---

// JoinRoom thêm connection vào một room nhất định
func (s *WSService) JoinRoom(roomID, userID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.rooms[roomID] == nil {
		s.rooms[roomID] = make(map[*websocket.Conn]string)
	}
	s.rooms[roomID][conn] = userID

	if s.connRooms[conn] == nil {
		s.connRooms[conn] = make(map[string]struct{})
	}
	s.connRooms[conn][roomID] = struct{}{}
}

// LeaveRoom loại bỏ connection khỏi một room
func (s *WSService) LeaveRoom(roomID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conns, ok := s.rooms[roomID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(s.rooms, roomID)
		}
	}

	if rooms, ok := s.connRooms[conn]; ok {
		delete(rooms, roomID)
		if len(rooms) == 0 {
			delete(s.connRooms, conn)
		}
	}
}

// LeaveAllRooms loại bỏ connection khỏi tất cả room đang tham gia
func (s *WSService) LeaveAllRooms(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rooms, ok := s.connRooms[conn]
	if !ok {
		return
	}
	for roomID := range rooms {
		if conns, ok := s.rooms[roomID]; ok {
			delete(conns, conn)
			if len(conns) == 0 {
				delete(s.rooms, roomID)
			}
		}
	}
	delete(s.connRooms, conn)
}

// BroadcastToRoom gửi message tới toàn bộ connection trong room
func (s *WSService) BroadcastToRoom(roomID, fromUserID, content string) {
	s.mu.RLock()
	connsMap, ok := s.rooms[roomID]
	if !ok {
		s.mu.RUnlock()
		return
	}
	// copy để tránh giữ lock khi ghi WS
	conns := make([]*websocket.Conn, 0, len(connsMap))
	for conn := range connsMap {
		conns = append(conns, conn)
	}
	s.mu.RUnlock()

	for _, conn := range conns {
		_ = conn.WriteJSON(map[string]string{
			"type":    "chat_message",
			"room":    roomID,
			"from":    fromUserID,
			"content": content,
		})
	}
}
