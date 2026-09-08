package handlers

import (
	"net/http"
	"testing"
	"time"

	"eventman/backend/internal/middleware"
	"eventman/backend/internal/models"
	"eventman/backend/internal/testutil"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func newReviewTestRouter(db *gorm.DB, userID uint) *gin.Engine {
	h := NewReviewHandler(db)
	r := gin.New()
	r.POST("/reviews", func(c *gin.Context) { c.Set(middleware.CtxUserID, userID); h.Create(c) })
	r.GET("/reviews/reviewable", func(c *gin.Context) { c.Set(middleware.CtxUserID, userID); h.Reviewable(c) })
	return r
}

func seedEndedEventWithSuccessTx(t *testing.T, db *gorm.DB, userID uint) models.Transaction {
	t.Helper()
	event := models.Event{
		OrganizerID: 2, CategoryID: 1, Title: "Past Event", Slug: "past-event",
		Description: "d", Location: "l", City: "c", TotalSeats: 10, AvailableSeats: 9,
		StartDate: time.Now().Add(-48 * time.Hour), EndDate: time.Now().Add(-24 * time.Hour),
		Status: models.EventPublished,
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	tx := models.Transaction{
		InvoiceNo: "INV-REVIEW-1", UserID: userID, EventID: event.ID, Quantity: 1,
		UnitPrice: 100000, Subtotal: 100000, TotalPrice: 100000, Status: models.TxSuccess,
	}
	if err := db.Create(&tx).Error; err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestCreateReview_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	tx := seedEndedEventWithSuccessTx(t, db, 1)
	r := newReviewTestRouter(db, 1)

	w := doJSON(r, http.MethodPost, "/reviews", map[string]any{
		"transaction_id": tx.ID, "rating": 5, "comment": "Great event!",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateReview_DuplicateRejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	tx := seedEndedEventWithSuccessTx(t, db, 1)
	r := newReviewTestRouter(db, 1)

	doJSON(r, http.MethodPost, "/reviews", map[string]any{"transaction_id": tx.ID, "rating": 4, "comment": "Nice"})
	w := doJSON(r, http.MethodPost, "/reviews", map[string]any{"transaction_id": tx.ID, "rating": 5, "comment": "again"})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for a duplicate review, got %d", w.Code)
	}
}

func TestCreateReview_RejectedForOtherUsersTransaction(t *testing.T) {
	db := testutil.NewTestDB(t)
	tx := seedEndedEventWithSuccessTx(t, db, 1)
	r := newReviewTestRouter(db, 2) // different user

	w := doJSON(r, http.MethodPost, "/reviews", map[string]any{"transaction_id": tx.ID, "rating": 5, "comment": "not mine"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when reviewing someone else's transaction, got %d", w.Code)
	}
}

func TestCreateReview_RejectedBeforeEventEnds(t *testing.T) {
	db := testutil.NewTestDB(t)
	event := models.Event{
		OrganizerID: 2, CategoryID: 1, Title: "Future Event", Slug: "future-event",
		Description: "d", Location: "l", City: "c", TotalSeats: 10, AvailableSeats: 9,
		StartDate: time.Now().Add(24 * time.Hour), EndDate: time.Now().Add(48 * time.Hour),
		Status: models.EventPublished,
	}
	db.Create(&event)
	tx := models.Transaction{
		InvoiceNo: "INV-REVIEW-2", UserID: 1, EventID: event.ID, Quantity: 1,
		UnitPrice: 100000, Subtotal: 100000, TotalPrice: 100000, Status: models.TxSuccess,
	}
	db.Create(&tx)

	r := newReviewTestRouter(db, 1)
	w := doJSON(r, http.MethodPost, "/reviews", map[string]any{"transaction_id": tx.ID, "rating": 5, "comment": "too early"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for reviewing before the event ends, got %d", w.Code)
	}
}
