package handlers

import (
	"net/http"
	"time"

	"eventman/backend/internal/dto"
	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReviewHandler struct {
	db *gorm.DB
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{db: db}
}

// Create lets a customer review an event they successfully attended
// (a SUCCESS transaction on an event that has already ended), once per transaction.
func (h *ReviewHandler) Create(c *gin.Context) {
	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.UserID(c)

	var tx models.Transaction
	if err := h.db.Preload("Event").
		Where("id = ? AND user_id = ? AND status = ?", req.TransactionID, userID, models.TxSuccess).
		First(&tx).Error; err != nil {
		utils.Error(c, http.StatusForbidden, "no completed transaction found for this order")
		return
	}
	if tx.Event == nil || time.Now().Before(tx.Event.EndDate) {
		utils.Error(c, http.StatusBadRequest, "you can only review an event after it has ended")
		return
	}

	var existing models.Review
	if err := h.db.Where("transaction_id = ?", tx.ID).First(&existing).Error; err == nil {
		utils.Error(c, http.StatusConflict, "you have already reviewed this order")
		return
	}

	review := models.Review{
		TransactionID: tx.ID,
		EventID:       tx.EventID,
		UserID:        userID,
		Rating:        req.Rating,
		Comment:       req.Comment,
	}
	if err := h.db.Create(&review).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to submit review")
		return
	}

	utils.Success(c, http.StatusCreated, review)
}

// Reviewable lists the caller's completed, ended, not-yet-reviewed orders.
func (h *ReviewHandler) Reviewable(c *gin.Context) {
	userID := middleware.UserID(c)

	txs := []models.Transaction{}
	h.db.Preload("Event").
		Joins("JOIN events ON events.id = transactions.event_id").
		Where("transactions.user_id = ? AND transactions.status = ? AND events.end_date < ?", userID, models.TxSuccess, time.Now()).
		Where("NOT EXISTS (SELECT 1 FROM reviews WHERE reviews.transaction_id = transactions.id)").
		Order("events.end_date DESC").
		Find(&txs)

	utils.Success(c, http.StatusOK, txs)
}
