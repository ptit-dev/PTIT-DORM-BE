package handlers

import (
	"Backend_Dorm_PTIT/repository"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	RoomRepo *repository.RoomRepository
}

func NewRoomHandler(roomRepo *repository.RoomRepository) *RoomHandler {
	return &RoomHandler{RoomRepo: roomRepo}
}

// GET /api/v1/rooms
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, err := h.RoomRepo.GetAllRooms(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rooms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

// GET /api/v1/rooms/:room_name/students
func (h *RoomHandler) ListStudentsInRoom(c *gin.Context) {
	room := c.Param("room_name")
	students, err := h.RoomRepo.GetStudentsByRoom(context.Background(), room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get students in room"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"students": students})
}
