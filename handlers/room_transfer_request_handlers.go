package handlers

import (
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomTransferRequestHandler struct {
	Repo *repository.RoomTransferRequestRepository
}

func NewRoomTransferRequestHandler(db *sql.DB) *RoomTransferRequestHandler {
	return &RoomTransferRequestHandler{
		Repo: repository.NewRoomTransferRequestRepository(db),
	}
}

func (h *RoomTransferRequestHandler) Create(c *gin.Context) {
	var req models.RoomTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = uuid.New().String()
	req.PeerConfirmStatus = "pending"
	req.ManagerConfirmStatus = "pending"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	if err := h.Repo.Create(context.Background(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *RoomTransferRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	req, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *RoomTransferRequestHandler) List(c *gin.Context) {
	reqs, err := h.Repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reqs)
}

func (h *RoomTransferRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.RoomTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id
	req.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *RoomTransferRequestHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Sinh viên phòng muốn chuyển xác nhận (accept/reject)
func (h *RoomTransferRequestHandler) PeerConfirm(c *gin.Context) {
	id := c.Param("id")
	type PeerConfirmInput struct {
		PeerConfirmStatus string `json:"peer_confirm_status" binding:"required"`
	}
	var input PeerConfirmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if req.PeerConfirmStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already confirmed"})
		return
	}
	req.PeerConfirmStatus = input.PeerConfirmStatus
	req.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// Quản lý xác nhận (accept/reject), nếu accept thì update hợp đồng 2 sinh viên đổi phòng
func (h *RoomTransferRequestHandler) ManagerConfirm(c *gin.Context) {
	id := c.Param("id")
	type ManagerConfirmInput struct {
		ManagerConfirmStatus string `json:"manager_confirm_status" binding:"required"`
	}
	var input ManagerConfirmInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if req.PeerConfirmStatus != "accepted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Peer must accept first"})
		return
	}
	if req.ManagerConfirmStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already confirmed"})
		return
	}
	req.ManagerConfirmStatus = input.ManagerConfirmStatus
	req.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if input.ManagerConfirmStatus == "accepted" {

		contractRepo := repository.NewContractRepository(h.Repo.DB)
		contracts1, err1 := contractRepo.GetContractByStudentID(context.Background(), req.RequesterUserID)
		contracts2, err2 := contractRepo.GetContractByStudentID(context.Background(), req.TargetUserID) // TargetUserID là student_id của sinh viên kia
		if err1 != nil || err2 != nil || len(contracts1) == 0 || len(contracts2) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot find contracts for both students"})
			return
		}

		contract1 := contracts1[0]
		contract2 := contracts2[0]
		room1 := contract1.Room
		room2 := contract2.Room

		errA := contractRepo.UpdateRoom(context.Background(), contract1.ID.String(), room2)
		errB := contractRepo.UpdateRoom(context.Background(), contract2.ID.String(), room1)
		if errA != nil || errB != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contracts room"})
			return
		}
	}
	c.JSON(http.StatusOK, req)
}
