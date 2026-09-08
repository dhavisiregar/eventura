package services

import (
	"testing"
	"time"

	"eventman/backend/internal/models"
	"eventman/backend/internal/testutil"

	"gorm.io/gorm"
)

func mustCreateUser(t *testing.T, db *gorm.DB, email string) models.User {
	t.Helper()
	u := models.User{Name: "Test User", Email: email, PasswordHash: "x", Role: models.RoleCustomer, ReferralCode: email}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return u
}

func TestAwardReferralPoints_CreditsBalance(t *testing.T) {
	db := testutil.NewTestDB(t)
	referrer := mustCreateUser(t, db, "referrer@test.com")

	if err := AwardReferralPoints(db, referrer.ID); err != nil {
		t.Fatalf("AwardReferralPoints returned error: %v", err)
	}

	balance, err := PointBalance(db, referrer.ID)
	if err != nil {
		t.Fatalf("PointBalance returned error: %v", err)
	}
	if balance != ReferralPointsAward {
		t.Errorf("expected balance %d, got %d", ReferralPointsAward, balance)
	}
}

func TestPointBalance_ExpiredPointsExcluded(t *testing.T) {
	db := testutil.NewTestDB(t)
	user := mustCreateUser(t, db, "expiry@test.com")

	past := time.Now().Add(-time.Hour)
	db.Create(&models.PointTransaction{UserID: user.ID, Points: 10000, Type: models.PointEarn, ExpiresAt: &past})

	balance, err := PointBalance(db, user.ID)
	if err != nil {
		t.Fatalf("PointBalance returned error: %v", err)
	}
	if balance != 0 {
		t.Errorf("expected expired points to be excluded, got balance %d", balance)
	}
}

func TestRedeemPoints_ReducesBalance(t *testing.T) {
	db := testutil.NewTestDB(t)
	user := mustCreateUser(t, db, "redeem@test.com")
	if err := AwardReferralPoints(db, user.ID); err != nil {
		t.Fatal(err)
	}

	if err := RedeemPoints(db, user.ID, 4000, "test redemption"); err != nil {
		t.Fatalf("RedeemPoints returned error: %v", err)
	}

	balance, _ := PointBalance(db, user.ID)
	if balance != ReferralPointsAward-4000 {
		t.Errorf("expected balance %d, got %d", ReferralPointsAward-4000, balance)
	}
}

func TestRefundPoints_RestoresBalance(t *testing.T) {
	db := testutil.NewTestDB(t)
	user := mustCreateUser(t, db, "refund@test.com")
	if err := AwardReferralPoints(db, user.ID); err != nil {
		t.Fatal(err)
	}
	if err := RedeemPoints(db, user.ID, 4000, "spend"); err != nil {
		t.Fatal(err)
	}
	if err := RefundPoints(db, user.ID, 4000, "refund"); err != nil {
		t.Fatalf("RefundPoints returned error: %v", err)
	}

	balance, _ := PointBalance(db, user.ID)
	if balance != ReferralPointsAward {
		t.Errorf("expected balance restored to %d, got %d", ReferralPointsAward, balance)
	}
}

func TestGrantReferralCoupon_ValidFor3Months(t *testing.T) {
	db := testutil.NewTestDB(t)
	user := mustCreateUser(t, db, "coupon@test.com")

	coupon, err := GrantReferralCoupon(db, user.ID)
	if err != nil {
		t.Fatalf("GrantReferralCoupon returned error: %v", err)
	}
	if coupon.DiscountValue != ReferralCouponDiscountPercent {
		t.Errorf("expected discount value %v, got %v", ReferralCouponDiscountPercent, coupon.DiscountValue)
	}

	expectedExpiry := time.Now().AddDate(0, CouponValidityMonths, 0)
	if coupon.ExpiresAt.Sub(expectedExpiry) > time.Minute || expectedExpiry.Sub(coupon.ExpiresAt) > time.Minute {
		t.Errorf("expected coupon to expire around %v, got %v", expectedExpiry, coupon.ExpiresAt)
	}
	if coupon.IsUsed {
		t.Error("a freshly granted coupon should not be marked used")
	}
}
