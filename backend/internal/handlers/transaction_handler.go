package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"eventman/backend/internal/dto"
	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/services"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	db  *gorm.DB
	txs *services.TransactionService
}

func NewTransactionHandler(db *gorm.DB, txs *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{db: db, txs: txs}
}

func (h *TransactionHandler) Checkout(c *gin.Context) {
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := h.db.First(&user, middleware.UserID(c)).Error; err != nil {
		utils.Error(c, http.StatusUnauthorized, "user not found")
		return
	}

	tx, err := h.txs.Checkout(&user, req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrEventNotAvailable) || errors.Is(err, services.ErrNotEnoughSeats) {
			status = http.StatusConflict
		}
		utils.Error(c, status, err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, tx)
}

func (h *TransactionHandler) MyTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	query := h.db.Model(&models.Transaction{}).Where("user_id = ?", middleware.UserID(c))
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	txs := []models.Transaction{}
	query.Preload("Event").Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).Find(&txs)

	utils.SuccessWithMeta(c, http.StatusOK, txs, utils.NewPagination(page, limit, total))
}

func (h *TransactionHandler) Detail(c *gin.Context) {
	var tx models.Transaction
	if err := h.db.Preload("Event").Where("id = ? AND user_id = ?", c.Param("id"), middleware.UserID(c)).
		First(&tx).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "transaction not found")
		return
	}
	utils.Success(c, http.StatusOK, tx)
}

func (h *TransactionHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.txs.CancelByUser(middleware.UserID(c), uint(id)); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, http.StatusOK, gin.H{"cancelled": true})
}

// SyncStatus lets the customer's own browser (after returning from the
// Midtrans payment page) ask us to reconcile the order's status immediately,
// instead of waiting for the server-to-server webhook.
func (h *TransactionHandler) SyncStatus(c *gin.Context) {
	tx, err := h.txs.SyncStatusForUser(middleware.UserID(c), c.Param("orderID"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, http.StatusOK, tx)
}

// MidtransNotification is Midtrans's server-to-server webhook. No auth
// middleware; authenticity is verified via the payload signature.
func (h *TransactionHandler) MidtransNotification(c *gin.Context, verify func(orderID, statusCode, grossAmount, signature string) bool) {
	var payload struct {
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		SignatureKey      string `json:"signature_key"`
		TransactionStatus string `json:"transaction_status"`
		FraudStatus       string `json:"fraud_status"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid payload")
		return
	}

	if !verify(payload.OrderID, payload.StatusCode, payload.GrossAmount, payload.SignatureKey) {
		utils.Error(c, http.StatusUnauthorized, "invalid signature")
		return
	}

	if err := h.txs.HandleMidtransNotification(payload.OrderID, payload.TransactionStatus, payload.FraudStatus); err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to process notification")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// --- Organizer ---

func (h *TransactionHandler) OrganizerTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	query := h.db.Model(&models.Transaction{}).
		Joins("JOIN events ON events.id = transactions.event_id").
		Where("events.organizer_id = ?", middleware.UserID(c))

	if eventID := c.Query("event_id"); eventID != "" {
		query = query.Where("transactions.event_id = ?", eventID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("transactions.status = ?", status)
	}

	var total int64
	query.Count(&total)

	txs := []models.Transaction{}
	query.Preload("Event").Preload("User").
		Order("transactions.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).Find(&txs)

	utils.SuccessWithMeta(c, http.StatusOK, txs, utils.NewPagination(page, limit, total))
}
