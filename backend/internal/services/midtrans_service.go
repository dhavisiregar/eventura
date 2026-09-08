package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"eventman/backend/internal/config"
)

type MidtransService struct {
	serverKey   string
	snapBaseURL string
	apiBaseURL  string
	frontendURL string
}

func NewMidtransService(cfg *config.Config) *MidtransService {
	snapBaseURL := "https://app.sandbox.midtrans.com/snap/v1"
	apiBaseURL := "https://api.sandbox.midtrans.com/v2"
	if cfg.MidtransIsProd {
		snapBaseURL = "https://app.midtrans.com/snap/v1"
		apiBaseURL = "https://api.midtrans.com/v2"
	}
	return &MidtransService{
		serverKey:   cfg.MidtransServerKey,
		snapBaseURL: snapBaseURL,
		apiBaseURL:  apiBaseURL,
		frontendURL: strings.TrimRight(cfg.FrontendURL, "/"),
	}
}

type SnapCustomer struct {
	FirstName string
	Email     string
}

type SnapResult struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

// CreateSnapTransaction requests a Snap payment page/token for the given order.
// grossAmount must be a whole-number IDR amount (Midtrans does not use decimals for IDR).
func (m *MidtransService) CreateSnapTransaction(orderID string, grossAmount int64, customer SnapCustomer) (*SnapResult, error) {
	if m.serverKey == "" {
		return nil, fmt.Errorf("midtrans server key is not configured")
	}

	// The browser is sent back here once payment finishes (success, pending, or
	// error). This is a plain client-side redirect, so localhost works fine for
	// local dev — unlike the server-to-server notification webhook below, which
	// Midtrans's servers cannot reach on localhost.
	returnURL := m.frontendURL + "/checkout/return"

	payload := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     orderID,
			"gross_amount": grossAmount,
		},
		"customer_details": map[string]any{
			"first_name": customer.FirstName,
			"email":      customer.Email,
		},
		"credit_card": map[string]any{"secure": true},
		"callbacks":   map[string]any{"finish": returnURL},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, m.snapBaseURL+"/transactions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(m.serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("midtrans error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result SnapResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifySignature validates the notification webhook signature per Midtrans docs:
// SHA512(order_id + status_code + gross_amount + ServerKey)
func (m *MidtransService) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	raw := orderID + statusCode + grossAmount + m.serverKey
	h := sha512.New()
	h.Write([]byte(raw))
	expected := hex.EncodeToString(h.Sum(nil))
	return expected == signatureKey
}

type StatusResult struct {
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

// GetTransactionStatus asks Midtrans directly for an order's current status.
// Used to reconcile a transaction from the client-side "finish" redirect,
// since the server-to-server notification webhook can't reach localhost
// during local development.
func (m *MidtransService) GetTransactionStatus(orderID string) (*StatusResult, error) {
	if m.serverKey == "" {
		return nil, fmt.Errorf("midtrans server key is not configured")
	}

	req, err := http.NewRequest(http.MethodGet, m.apiBaseURL+"/"+orderID+"/status", nil)
	if err != nil {
		return nil, err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(m.serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("midtrans status error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result StatusResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
