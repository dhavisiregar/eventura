package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eventman/backend/internal/config"
	"eventman/backend/internal/dto"
	"eventman/backend/internal/models"
	"eventman/backend/internal/services"
	"eventman/backend/internal/testutil"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newAuthTestRouter(db *gorm.DB) *gin.Engine {
	cfg := &config.Config{JWTSecret: "test-secret"}
	h := NewAuthHandler(db, cfg)

	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	return r
}

func doJSON(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegister_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newAuthTestRouter(db)

	w := doJSON(r, http.MethodPost, "/register", dto.RegisterRequest{
		Name: "Alice", Email: "alice@test.com", Password: "password1", Role: "customer",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var user models.User
	if err := db.Where("email = ?", "alice@test.com").First(&user).Error; err != nil {
		t.Fatalf("expected user to be persisted: %v", err)
	}
	if user.ReferralCode == "" {
		t.Error("expected a referral code to be generated")
	}
}

func TestRegister_DuplicateEmailRejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newAuthTestRouter(db)

	body := dto.RegisterRequest{Name: "Alice", Email: "dupe@test.com", Password: "password1", Role: "customer"}
	doJSON(r, http.MethodPost, "/register", body)
	w := doJSON(r, http.MethodPost, "/register", body)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate email, got %d", w.Code)
	}
}

func TestRegister_WithReferralCode_AwardsPointsAndCoupon(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newAuthTestRouter(db)

	doJSON(r, http.MethodPost, "/register", dto.RegisterRequest{
		Name: "Referrer", Email: "referrer@test.com", Password: "password1", Role: "customer",
	})
	var referrer models.User
	db.Where("email = ?", "referrer@test.com").First(&referrer)

	w := doJSON(r, http.MethodPost, "/register", dto.RegisterRequest{
		Name: "Referred", Email: "referred@test.com", Password: "password1", Role: "customer",
		ReferralCode: referrer.ReferralCode,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	balance, _ := services.PointBalance(db, referrer.ID)
	if balance != services.ReferralPointsAward {
		t.Errorf("expected referrer to receive %d points, got %d", services.ReferralPointsAward, balance)
	}

	var referred models.User
	db.Where("email = ?", "referred@test.com").First(&referred)
	var coupon models.Coupon
	if err := db.Where("user_id = ?", referred.ID).First(&coupon).Error; err != nil {
		t.Fatalf("expected referred user to receive a welcome coupon: %v", err)
	}
	if coupon.DiscountValue != services.ReferralCouponDiscountPercent {
		t.Errorf("expected 10%% discount coupon, got %v", coupon.DiscountValue)
	}
}

func TestRegister_InvalidReferralCodeRejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newAuthTestRouter(db)

	w := doJSON(r, http.MethodPost, "/register", dto.RegisterRequest{
		Name: "Bob", Email: "bob@test.com", Password: "password1", Role: "customer",
		ReferralCode: "DOES-NOT-EXIST",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown referral code, got %d", w.Code)
	}
}

func TestLogin_WrongPasswordRejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newAuthTestRouter(db)

	doJSON(r, http.MethodPost, "/register", dto.RegisterRequest{
		Name: "Carol", Email: "carol@test.com", Password: "correct-password", Role: "customer",
	})

	w := doJSON(r, http.MethodPost, "/login", dto.LoginRequest{Email: "carol@test.com", Password: "wrong-password"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	r := newAuthTestRouter(db)

	doJSON(r, http.MethodPost, "/register", dto.RegisterRequest{
		Name: "Dan", Email: "dan@test.com", Password: "correct-password", Role: "organizer",
	})

	w := doJSON(r, http.MethodPost, "/login", dto.LoginRequest{Email: "dan@test.com", Password: "correct-password"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data dto.AuthResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Data.Token == "" {
		t.Error("expected a JWT token in the login response")
	}
}
