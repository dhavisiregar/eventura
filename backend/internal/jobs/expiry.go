package jobs

import (
	"log"
	"time"

	"eventman/backend/internal/services"
)

// StartTransactionExpirySweeper periodically releases seats/discounts/points
// held by pending orders whose payment deadline has passed.
func StartTransactionExpirySweeper(svc *services.TransactionService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			n, err := svc.SweepExpiredTransactions()
			if err != nil {
				log.Printf("expiry sweep error: %v", err)
				continue
			}
			if n > 0 {
				log.Printf("expiry sweep: expired %d pending transaction(s)", n)
			}
		}
	}()
}
