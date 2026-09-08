package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all environment-driven configuration for the API.
type Config struct {
	AppEnv          string
	Port            string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSL           bool
	JWTSecret       string
	JWTExpiry       time.Duration
	FrontendURL     string
	UploadDir       string
	PaymentDeadline time.Duration

	MidtransServerKey string
	MidtransClientKey string
	MidtransIsProd    bool
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// Load reads configuration from environment variables (populated via .env in dev).
func Load() *Config {
	jwtExpiryHours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))
	if err != nil {
		jwtExpiryHours = 72
	}
	paymentDeadlineMinutes, err := strconv.Atoi(getEnv("PAYMENT_DEADLINE_MINUTES", "120"))
	if err != nil {
		paymentDeadlineMinutes = 120
	}

	return &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "127.0.0.1"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBUser:          getEnv("DB_USER", "root"),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "event_management"),
		DBSSL:           getEnv("DB_SSL", "false") == "true",
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiry:       time.Duration(jwtExpiryHours) * time.Hour,
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		PaymentDeadline: time.Duration(paymentDeadlineMinutes) * time.Minute,

		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransIsProd:    getEnv("MIDTRANS_IS_PRODUCTION", "false") == "true",
	}
}
