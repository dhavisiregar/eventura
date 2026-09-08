package services

import (
	"errors"
	"fmt"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/dto"
	"eventman/backend/internal/models"
	"eventman/backend/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionService struct {
	db       *gorm.DB
	midtrans *MidtransService
	cfg      *config.Config
}

func NewTransactionService(db *gorm.DB, midtrans *MidtransService, cfg *config.Config) *TransactionService {
	return &TransactionService{db: db, midtrans: midtrans, cfg: cfg}
}

var (
	ErrEventNotAvailable  = errors.New("event is not available for booking")
	ErrNotEnoughSeats     = errors.New("not enough seats available")
	ErrInvalidVoucher     = errors.New("voucher code is invalid, expired, or fully redeemed")
	ErrInvalidCoupon      = errors.New("no valid referral coupon available")
	ErrInsufficientPoints = errors.New("insufficient point balance")
)

// Checkout creates a transaction, reserves seats, applies discounts, and (for
// paid orders) opens a Midtrans Snap payment. Every step that mutates rows
// runs inside a single DB transaction so a failure anywhere rolls everything
// back atomically.
func (s *TransactionService) Checkout(user *models.User, req dto.CreateTransactionRequest) (*models.Transaction, error) {
	var result *models.Transaction

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var event models.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, req.EventID).Error; err != nil {
			return ErrEventNotAvailable
		}
		if event.Status != models.EventPublished || time.Now().After(event.StartDate) {
			return ErrEventNotAvailable
		}

		unitPrice := 0.0
		var ticketType *models.TicketType
		if req.TicketTypeID != nil {
			var tt models.TicketType
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? AND event_id = ?", *req.TicketTypeID, event.ID).First(&tt).Error; err != nil {
				return errors.New("ticket type not found")
			}
			if tt.Remaining < req.Quantity {
				return ErrNotEnoughSeats
			}
			unitPrice = tt.Price
			ticketType = &tt
		} else {
			if event.AvailableSeats < req.Quantity {
				return ErrNotEnoughSeats
			}
			if event.IsPaid {
				unitPrice = event.Price
			}
		}

		subtotal := unitPrice * float64(req.Quantity)

		voucherDiscount := 0.0
		var voucher *models.Voucher
		if req.VoucherCode != "" {
			var v models.Voucher
			now := time.Now()
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("event_id = ? AND code = ? AND start_date <= ? AND end_date >= ? AND used_count < quota",
					event.ID, req.VoucherCode, now, now).
				First(&v).Error
			if err != nil {
				return ErrInvalidVoucher
			}
			if v.DiscountType == models.DiscountPercent {
				voucherDiscount = subtotal * v.DiscountValue / 100
			} else {
				voucherDiscount = v.DiscountValue
			}
			if voucherDiscount > subtotal {
				voucherDiscount = subtotal
			}
			voucher = &v
		}

		remainingAfterVoucher := subtotal - voucherDiscount
		couponDiscount := 0.0
		var coupon *models.Coupon
		if req.UseCoupon {
			var c models.Coupon
			now := time.Now()
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ? AND is_used = ? AND expires_at > ? AND source = ?", user.ID, false, now, "referral_signup").
				Order("expires_at ASC").First(&c).Error
			if err != nil {
				return ErrInvalidCoupon
			}
			couponDiscount = remainingAfterVoucher * c.DiscountValue / 100
			coupon = &c
		}

		remainingAfterCoupon := remainingAfterVoucher - couponDiscount
		pointsDiscount := 0.0
		pointsUsed := 0
		if req.PointsToUse > 0 {
			balance, err := PointBalance(tx, user.ID)
			if err != nil {
				return err
			}
			if req.PointsToUse > balance {
				return ErrInsufficientPoints
			}
			pointsUsed = req.PointsToUse
			pointsDiscount = float64(pointsUsed)
			if pointsDiscount > remainingAfterCoupon {
				pointsDiscount = remainingAfterCoupon
				pointsUsed = int(remainingAfterCoupon)
			}
		}

		total := remainingAfterCoupon - pointsDiscount
		if total < 0 {
			total = 0
		}

		// Reserve seats.
		if ticketType != nil {
			ticketType.Remaining -= req.Quantity
			if err := tx.Model(&models.TicketType{}).Where("id = ?", ticketType.ID).
				Update("remaining", ticketType.Remaining).Error; err != nil {
				return err
			}
		}
		event.AvailableSeats -= req.Quantity
		if err := tx.Model(&models.Event{}).Where("id = ?", event.ID).
			Update("available_seats", event.AvailableSeats).Error; err != nil {
			return err
		}

		if voucher != nil {
			if err := tx.Model(&models.Voucher{}).Where("id = ?", voucher.ID).
				Update("used_count", voucher.UsedCount+1).Error; err != nil {
				return err
			}
		}
		if coupon != nil {
			now := time.Now()
			if err := tx.Model(&models.Coupon{}).Where("id = ?", coupon.ID).
				Updates(map[string]any{"is_used": true, "used_at": &now}).Error; err != nil {
				return err
			}
		}
		if pointsUsed > 0 {
			if err := RedeemPoints(tx, user.ID, pointsUsed, "Used for ticket purchase"); err != nil {
				return err
			}
		}

		txRecord := models.Transaction{
			InvoiceNo:       utils.GenerateInvoiceNo(),
			UserID:          user.ID,
			EventID:         event.ID,
			TicketTypeID:    req.TicketTypeID,
			Quantity:        req.Quantity,
			UnitPrice:       unitPrice,
			Subtotal:        subtotal,
			VoucherDiscount: voucherDiscount,
			CouponDiscount:  couponDiscount,
			PointsUsed:      pointsUsed,
			PointsDiscount:  pointsDiscount,
			TotalPrice:      total,
			PaymentDeadline: time.Now().Add(s.cfg.PaymentDeadline),
		}
		if voucher != nil {
			txRecord.VoucherID = &voucher.ID
		}
		if coupon != nil {
			txRecord.CouponID = &coupon.ID
		}

		if total <= 0 {
			// Free event, or fully covered by discounts/points: no payment needed.
			txRecord.Status = models.TxSuccess
		} else {
			txRecord.Status = models.TxPendingPayment
			orderID := utils.GenerateOrderID()
			snap, err := s.midtrans.CreateSnapTransaction(orderID, int64(total), SnapCustomer{
				FirstName: user.Name,
				Email:     user.Email,
			})
			if err != nil {
				return fmt.Errorf("payment gateway error: %w", err)
			}
			txRecord.MidtransOrderID = orderID
			txRecord.MidtransToken = snap.Token
			txRecord.MidtransRedirectURL = snap.RedirectURL
		}

		if err := tx.Create(&txRecord).Error; err != nil {
			return err
		}
		result = &txRecord
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// releaseReservation reverses seat/voucher/coupon/points holds for a transaction
// that failed, expired, or was cancelled. Must run inside a DB transaction.
func releaseReservation(tx *gorm.DB, t *models.Transaction) error {
	if t.TicketTypeID != nil {
		if err := tx.Model(&models.TicketType{}).Where("id = ?", *t.TicketTypeID).
			Update("remaining", gorm.Expr("remaining + ?", t.Quantity)).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&models.Event{}).Where("id = ?", t.EventID).
		Update("available_seats", gorm.Expr("available_seats + ?", t.Quantity)).Error; err != nil {
		return err
	}
	if t.VoucherID != nil {
		if err := tx.Model(&models.Voucher{}).Where("id = ?", *t.VoucherID).
			Update("used_count", gorm.Expr("used_count - 1")).Error; err != nil {
			return err
		}
	}
	if t.CouponID != nil {
		if err := tx.Model(&models.Coupon{}).Where("id = ?", *t.CouponID).
			Updates(map[string]any{"is_used": false, "used_at": nil}).Error; err != nil {
			return err
		}
	}
	if t.PointsUsed > 0 {
		if err := RefundPoints(tx, t.UserID, t.PointsUsed, "Refund: order "+t.InvoiceNo+" not completed"); err != nil {
			return err
		}
	}
	return nil
}

// HandleMidtransNotification updates a transaction's status from a Midtrans
// webhook payload and releases reserved resources on failure.
func (s *TransactionService) HandleMidtransNotification(orderID, transactionStatus, fraudStatus string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var t models.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("midtrans_order_id = ?", orderID).First(&t).Error; err != nil {
			return err
		}
		if t.Status != models.TxPendingPayment {
			return nil // already finalized, ignore duplicate notifications
		}

		switch transactionStatus {
		case "capture", "settlement":
			if fraudStatus == "" || fraudStatus == "accept" {
				t.Status = models.TxSuccess
			}
		case "pending":
			return nil
		case "deny", "cancel", "expire", "failure":
			t.Status = models.TxFailed
			if transactionStatus == "expire" {
				t.Status = models.TxExpired
			}
			if err := releaseReservation(tx, &t); err != nil {
				return err
			}
		}
		return tx.Model(&models.Transaction{}).Where("id = ?", t.ID).Update("status", t.Status).Error
	})
}

// SyncStatusForUser lets the customer's own browser reconcile a transaction's
// status right after returning from Midtrans's payment page, by asking
// Midtrans directly for the authoritative status. This covers local
// development, where Midtrans's server-to-server notification webhook has no
// way to reach a "localhost" backend; in a deployment with a public webhook
// URL configured, this is a (harmless, idempotent) backup path.
func (s *TransactionService) SyncStatusForUser(userID uint, orderID string) (*models.Transaction, error) {
	var t models.Transaction
	if err := s.db.Where("midtrans_order_id = ? AND user_id = ?", orderID, userID).First(&t).Error; err != nil {
		return nil, errors.New("transaction not found")
	}

	if t.Status == models.TxPendingPayment {
		status, err := s.midtrans.GetTransactionStatus(orderID)
		if err != nil {
			return nil, fmt.Errorf("failed to check payment status: %w", err)
		}
		if err := s.HandleMidtransNotification(orderID, status.TransactionStatus, status.FraudStatus); err != nil {
			return nil, err
		}
		if err := s.db.Where("midtrans_order_id = ?", orderID).First(&t).Error; err != nil {
			return nil, err
		}
	}

	return &t, nil
}

// CancelByUser lets a customer cancel their own pending order.
func (s *TransactionService) CancelByUser(userID, transactionID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var t models.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", transactionID, userID).First(&t).Error; err != nil {
			return errors.New("transaction not found")
		}
		if t.Status != models.TxPendingPayment {
			return errors.New("only pending transactions can be cancelled")
		}
		t.Status = models.TxCancelled
		if err := releaseReservation(tx, &t); err != nil {
			return err
		}
		return tx.Model(&models.Transaction{}).Where("id = ?", t.ID).Update("status", t.Status).Error
	})
}

// SweepExpiredTransactions finds pending orders past their payment deadline,
// marks them expired, and releases whatever they were holding. Intended to
// run on a periodic background ticker.
func (s *TransactionService) SweepExpiredTransactions() (int, error) {
	var stale []models.Transaction
	if err := s.db.Where("status = ? AND payment_deadline < ?", models.TxPendingPayment, time.Now()).
		Find(&stale).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, t := range stale {
		err := s.db.Transaction(func(tx *gorm.DB) error {
			var locked models.Transaction
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, t.ID).Error; err != nil {
				return err
			}
			if locked.Status != models.TxPendingPayment {
				return nil
			}
			if err := releaseReservation(tx, &locked); err != nil {
				return err
			}
			return tx.Model(&models.Transaction{}).Where("id = ?", locked.ID).
				Update("status", models.TxExpired).Error
		})
		if err == nil {
			count++
		}
	}
	return count, nil
}
