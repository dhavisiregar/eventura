package handlers

import (
	"net/http"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/dto"
	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/services"
	"eventman/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var existing models.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		utils.Error(c, http.StatusConflict, "email is already registered")
		return
	}

	var referrer *models.User
	if req.ReferralCode != "" {
		var r models.User
		if err := h.db.Where("referral_code = ?", req.ReferralCode).First(&r).Error; err != nil {
			utils.Error(c, http.StatusBadRequest, "referral code not found")
			return
		}
		referrer = &r
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to process password")
		return
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         models.Role(req.Role),
		ReferralCode: utils.GenerateReferralCode(),
	}
	if referrer != nil {
		user.ReferredBy = &referrer.ID
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if referrer != nil {
			if err := services.AwardReferralPoints(tx, referrer.ID); err != nil {
				return err
			}
			if _, err := services.GrantReferralCoupon(tx, user.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to register: "+err.Error())
		return
	}

	token, err := utils.GenerateToken(h.cfg.JWTSecret, h.cfg.JWTExpiry, user.ID, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.Success(c, http.StatusCreated, dto.AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		utils.Error(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := utils.GenerateToken(h.cfg.JWTSecret, h.cfg.JWTExpiry, user.ID, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.Success(c, http.StatusOK, dto.AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, middleware.UserID(c)).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}
	utils.Success(c, http.StatusOK, user)
}

func (h *AuthHandler) ReferralInfo(c *gin.Context) {
	userID := middleware.UserID(c)

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	var referralCount int64
	h.db.Model(&models.User{}).Where("referred_by = ?", userID).Count(&referralCount)

	balance, _ := services.PointBalance(h.db, userID)

	coupons := []models.Coupon{}
	h.db.Where("user_id = ? AND is_used = ? AND expires_at > ?", userID, false, time.Now()).Find(&coupons)

	ledger := []models.PointTransaction{}
	h.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&ledger)

	utils.Success(c, http.StatusOK, gin.H{
		"referral_code":     user.ReferralCode,
		"total_referrals":   referralCount,
		"point_balance":     balance,
		"available_coupons": coupons,
		"point_ledger":      ledger,
	})
}
