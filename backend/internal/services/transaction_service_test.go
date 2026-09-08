package services

import (
	"testing"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/dto"
	"eventman/backend/internal/models"
	"eventman/backend/internal/testutil"

	"gorm.io/gorm"
)

func newTestTxService(t *testing.T, db *gorm.DB) *TransactionService {
	t.Helper()
	cfg := &config.Config{PaymentDeadline: 2 * time.Hour} // MidtransServerKey left empty on purpose
	return NewTransactionService(db, NewMidtransService(cfg), cfg)
}

func mustCreateBuyer(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	u := models.User{Name: "Buyer", Email: "buyer@test.com", PasswordHash: "x", Role: models.RoleCustomer, ReferralCode: "BUYER1"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	return u
}

func mustCreateEvent(t *testing.T, db *gorm.DB, isPaid bool, price float64, seats int) models.Event {
	t.Helper()
	e := models.Event{
		OrganizerID: 1, CategoryID: 1, Title: "Test Event", Slug: "test-event-" + time.Now().Format("150405.000000"),
		Description: "desc", Location: "loc", City: "city",
		IsPaid: isPaid, Price: price, TotalSeats: seats, AvailableSeats: seats,
		StartDate: time.Now().Add(24 * time.Hour), EndDate: time.Now().Add(48 * time.Hour),
		Status: models.EventPublished,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

func TestCheckout_FreeEvent_SucceedsImmediately(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db)
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, false, 0, 10)

	tx, err := svc.Checkout(&buyer, dto.CreateTransactionRequest{EventID: event.ID, Quantity: 2})
	if err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}
	if tx.Status != models.TxSuccess {
		t.Errorf("expected free event order to be immediately successful, got status %q", tx.Status)
	}

	var updated models.Event
	db.First(&updated, event.ID)
	if updated.AvailableSeats != 8 {
		t.Errorf("expected 8 seats remaining, got %d", updated.AvailableSeats)
	}
}

func TestCheckout_NotEnoughSeats(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db)
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, false, 0, 1)

	_, err := svc.Checkout(&buyer, dto.CreateTransactionRequest{EventID: event.ID, Quantity: 5})
	if err != ErrNotEnoughSeats {
		t.Fatalf("expected ErrNotEnoughSeats, got %v", err)
	}

	var updated models.Event
	db.First(&updated, event.ID)
	if updated.AvailableSeats != 1 {
		t.Errorf("seat count should be unchanged after a failed checkout, got %d", updated.AvailableSeats)
	}
}

func TestCheckout_PaidEvent_GatewayFailureRollsBackSeats(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db) // Midtrans server key intentionally unset
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, true, 100000, 10)

	_, err := svc.Checkout(&buyer, dto.CreateTransactionRequest{EventID: event.ID, Quantity: 1})
	if err == nil {
		t.Fatal("expected checkout to fail without a configured payment gateway")
	}

	var updated models.Event
	db.First(&updated, event.ID)
	if updated.AvailableSeats != 10 {
		t.Errorf("seats should roll back on gateway failure, got %d available", updated.AvailableSeats)
	}
}

func TestCheckout_PointsFullyCoverPrice_SkipsPaymentGateway(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db)
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, true, 5000, 10)

	if err := AwardReferralPoints(db, buyer.ID); err != nil { // 10,000 points
		t.Fatal(err)
	}

	tx, err := svc.Checkout(&buyer, dto.CreateTransactionRequest{EventID: event.ID, Quantity: 1, PointsToUse: 5000})
	if err != nil {
		t.Fatalf("Checkout returned error: %v", err)
	}
	if tx.Status != models.TxSuccess {
		t.Errorf("expected order fully covered by points to succeed immediately, got %q", tx.Status)
	}
	if tx.TotalPrice != 0 {
		t.Errorf("expected total price 0, got %v", tx.TotalPrice)
	}

	balance, _ := PointBalance(db, buyer.ID)
	if balance != 5000 {
		t.Errorf("expected remaining balance 5000, got %d", balance)
	}
}

func TestSyncStatusForUser_AlreadyFinalizedSkipsMidtransCall(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db) // Midtrans server key unset: a live API call would error out
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, true, 100000, 10)

	txRecord := models.Transaction{
		InvoiceNo: "INV-SYNC-1", UserID: buyer.ID, EventID: event.ID, Quantity: 1,
		UnitPrice: 100000, Subtotal: 100000, TotalPrice: 100000,
		Status: models.TxSuccess, MidtransOrderID: "ORDER-SYNC-1",
	}
	db.Create(&txRecord)

	got, err := svc.SyncStatusForUser(buyer.ID, "ORDER-SYNC-1")
	if err != nil {
		t.Fatalf("SyncStatusForUser returned error: %v", err)
	}
	if got.Status != models.TxSuccess {
		t.Errorf("expected already-finalized status to be returned unchanged, got %q", got.Status)
	}
}

func TestSyncStatusForUser_WrongUserRejected(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db)
	buyer := mustCreateBuyer(t, db)
	otherUser := mustCreateUser(t, db, "other@test.com")
	event := mustCreateEvent(t, db, true, 100000, 10)

	db.Create(&models.Transaction{
		InvoiceNo: "INV-SYNC-2", UserID: buyer.ID, EventID: event.ID, Quantity: 1,
		UnitPrice: 100000, Subtotal: 100000, TotalPrice: 100000,
		Status: models.TxSuccess, MidtransOrderID: "ORDER-SYNC-2",
	})

	if _, err := svc.SyncStatusForUser(otherUser.ID, "ORDER-SYNC-2"); err == nil {
		t.Error("expected an error when syncing an order that belongs to a different user")
	}
}

func TestCancelByUser_ReleasesSeats(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db)
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, true, 100000, 10)

	txRecord := models.Transaction{
		InvoiceNo: "INV-TEST-1", UserID: buyer.ID, EventID: event.ID, Quantity: 2,
		UnitPrice: 100000, Subtotal: 200000, TotalPrice: 200000,
		Status: models.TxPendingPayment, MidtransOrderID: "ORDER-TEST-1",
		PaymentDeadline: time.Now().Add(time.Hour),
	}
	db.Create(&txRecord)
	db.Model(&models.Event{}).Where("id = ?", event.ID).Update("available_seats", 8)

	if err := svc.CancelByUser(buyer.ID, txRecord.ID); err != nil {
		t.Fatalf("CancelByUser returned error: %v", err)
	}

	var updatedTx models.Transaction
	db.First(&updatedTx, txRecord.ID)
	if updatedTx.Status != models.TxCancelled {
		t.Errorf("expected status cancelled, got %q", updatedTx.Status)
	}

	var updatedEvent models.Event
	db.First(&updatedEvent, event.ID)
	if updatedEvent.AvailableSeats != 10 {
		t.Errorf("expected seats restored to 10, got %d", updatedEvent.AvailableSeats)
	}
}

func TestSweepExpiredTransactions_ExpiresStaleOrders(t *testing.T) {
	db := testutil.NewTestDB(t)
	svc := newTestTxService(t, db)
	buyer := mustCreateBuyer(t, db)
	event := mustCreateEvent(t, db, true, 100000, 10)

	txRecord := models.Transaction{
		InvoiceNo: "INV-TEST-2", UserID: buyer.ID, EventID: event.ID, Quantity: 3,
		UnitPrice: 100000, Subtotal: 300000, TotalPrice: 300000,
		Status: models.TxPendingPayment, MidtransOrderID: "ORDER-TEST-2",
		PaymentDeadline: time.Now().Add(-time.Minute), // already past deadline
	}
	db.Create(&txRecord)
	db.Model(&models.Event{}).Where("id = ?", event.ID).Update("available_seats", 7)

	n, err := svc.SweepExpiredTransactions()
	if err != nil {
		t.Fatalf("SweepExpiredTransactions returned error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 expired transaction, got %d", n)
	}

	var updatedTx models.Transaction
	db.First(&updatedTx, txRecord.ID)
	if updatedTx.Status != models.TxExpired {
		t.Errorf("expected status expired, got %q", updatedTx.Status)
	}

	var updatedEvent models.Event
	db.First(&updatedEvent, event.ID)
	if updatedEvent.AvailableSeats != 10 {
		t.Errorf("expected seats restored to 10, got %d", updatedEvent.AvailableSeats)
	}
}
