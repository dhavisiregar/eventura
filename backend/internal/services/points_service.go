package services

import (
	"time"

	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"gorm.io/gorm"
)

const (
	ReferralPointsAward           = 10000
	PointsValidityMonths          = 3
	ReferralCouponDiscountPercent = 10
	CouponValidityMonths          = 3
)

// AwardReferralPoints credits the referrer with points after a referred signup.
// Points expire 3 months from now. Must run inside the same DB transaction
// as the new user's creation.
func AwardReferralPoints(tx *gorm.DB, referrerID uint) error {
	expires := time.Now().AddDate(0, PointsValidityMonths, 0)
	pt := models.PointTransaction{
		UserID:      referrerID,
		Points:      ReferralPointsAward,
		Type:        models.PointEarn,
		Description: "Referral signup bonus",
		ExpiresAt:   &expires,
	}
	return tx.Create(&pt).Error
}

// GrantReferralCoupon creates the 10% discount coupon for a newly-referred user,
// valid for 3 months.
func GrantReferralCoupon(tx *gorm.DB, userID uint) (*models.Coupon, error) {
	coupon := models.Coupon{
		UserID:        userID,
		Code:          "WELCOME-" + utils.RandomCode(6),
		DiscountType:  models.DiscountPercent,
		DiscountValue: ReferralCouponDiscountPercent,
		Source:        "referral_signup",
		ExpiresAt:     time.Now().AddDate(0, CouponValidityMonths, 0),
	}
	if err := tx.Create(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// PointBalance is the usable point balance: unexpired earned points minus all redemptions.
func PointBalance(db *gorm.DB, userID uint) (int, error) {
	var earned, redeemed int64
	now := time.Now()

	if err := db.Model(&models.PointTransaction{}).
		Where("user_id = ? AND type = ? AND (expires_at IS NULL OR expires_at > ?)", userID, models.PointEarn, now).
		Select("COALESCE(SUM(points),0)").Scan(&earned).Error; err != nil {
		return 0, err
	}
	if err := db.Model(&models.PointTransaction{}).
		Where("user_id = ? AND type = ?", userID, models.PointRedeem).
		Select("COALESCE(SUM(points),0)").Scan(&redeemed).Error; err != nil {
		return 0, err
	}

	balance := int(earned) - int(redeemed)
	if balance < 0 {
		balance = 0
	}
	return balance, nil
}

// RedeemPoints records a redemption within an existing transaction (caller must
// have already validated the user has sufficient balance).
func RedeemPoints(tx *gorm.DB, userID uint, points int, description string) error {
	if points <= 0 {
		return nil
	}
	pt := models.PointTransaction{
		UserID:      userID,
		Points:      points,
		Type:        models.PointRedeem,
		Description: description,
	}
	return tx.Create(&pt).Error
}

// RefundPoints reverses a prior redemption (e.g. when an order is cancelled/expired)
// by crediting the points back with a fresh 3-month expiry.
func RefundPoints(tx *gorm.DB, userID uint, points int, description string) error {
	if points <= 0 {
		return nil
	}
	expires := time.Now().AddDate(0, PointsValidityMonths, 0)
	pt := models.PointTransaction{
		UserID:      userID,
		Points:      points,
		Type:        models.PointEarn,
		Description: description,
		ExpiresAt:   &expires,
	}
	return tx.Create(&pt).Error
}
