package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const alphanumeric = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no ambiguous 0/O/1/I

// RandomCode generates a cryptographically random uppercase alphanumeric code.
func RandomCode(length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphanumeric))))
		sb.WriteByte(alphanumeric[idx.Int64()])
	}
	return sb.String()
}

// GenerateReferralCode builds a short, user-friendly referral code.
func GenerateReferralCode() string {
	return RandomCode(8)
}

// GenerateInvoiceNo builds a human-readable, time-sortable invoice number.
func GenerateInvoiceNo() string {
	return fmt.Sprintf("INV-%s-%s", time.Now().Format("20060102150405"), RandomCode(4))
}

// GenerateOrderID builds a unique order id used as the Midtrans transaction identifier.
func GenerateOrderID() string {
	return fmt.Sprintf("ORDER-%s-%s", time.Now().Format("20060102150405"), RandomCode(6))
}
