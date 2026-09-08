package handlers

import (
	"net/http"

	"eventman/backend/internal/dto"
	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VoucherHandler struct {
	db *gorm.DB
}

func NewVoucherHandler(db *gorm.DB) *VoucherHandler {
	return &VoucherHandler{db: db}
}

func (h *VoucherHandler) ownsEvent(c *gin.Context, eventID any) (*models.Event, bool) {
	var event models.Event
	if err := h.db.First(&event, eventID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "event not found")
		return nil, false
	}
	if event.OrganizerID != middleware.UserID(c) {
		utils.Error(c, http.StatusForbidden, "you do not own this event")
		return nil, false
	}
	return &event, true
}

func (h *VoucherHandler) Create(c *gin.Context) {
	eventID := c.Param("id")
	if _, ok := h.ownsEvent(c, eventID); !ok {
		return
	}

	var req dto.CreateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.EndDate.Before(req.StartDate) {
		utils.Error(c, http.StatusBadRequest, "end_date must be after start_date")
		return
	}

	voucher := models.Voucher{
		EventID:       parseUint(eventID),
		Code:          req.Code,
		DiscountType:  models.DiscountType(req.DiscountType),
		DiscountValue: req.DiscountValue,
		Quota:         req.Quota,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
	}

	if err := h.db.Create(&voucher).Error; err != nil {
		utils.Error(c, http.StatusConflict, "a voucher with this code already exists for this event")
		return
	}
	utils.Success(c, http.StatusCreated, voucher)
}

func (h *VoucherHandler) List(c *gin.Context) {
	eventID := c.Param("id")
	if _, ok := h.ownsEvent(c, eventID); !ok {
		return
	}

	vouchers := []models.Voucher{}
	h.db.Where("event_id = ?", eventID).Order("created_at DESC").Find(&vouchers)
	utils.Success(c, http.StatusOK, vouchers)
}

func (h *VoucherHandler) Delete(c *gin.Context) {
	var voucher models.Voucher
	if err := h.db.First(&voucher, c.Param("voucherId")).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "voucher not found")
		return
	}
	if _, ok := h.ownsEvent(c, voucher.EventID); !ok {
		return
	}
	if voucher.UsedCount > 0 {
		utils.Error(c, http.StatusConflict, "cannot delete a voucher that has already been redeemed")
		return
	}
	h.db.Delete(&voucher)
	utils.Success(c, http.StatusOK, gin.H{"deleted": true})
}
