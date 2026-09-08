package utils

import (
	"testing"
	"time"

	"eventman/backend/internal/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "s3cret123" {
		t.Fatal("password was not hashed")
	}
	if !CheckPassword(hash, "s3cret123") {
		t.Error("CheckPassword should accept the correct password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("CheckPassword should reject an incorrect password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateToken(secret, time.Hour, 42, models.RoleOrganizer)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", claims.UserID)
	}
	if claims.Role != models.RoleOrganizer {
		t.Errorf("expected role organizer, got %s", claims.Role)
	}
}

func TestParseToken_WrongSecretRejected(t *testing.T) {
	token, _ := GenerateToken("secret-a", time.Hour, 1, models.RoleCustomer)
	if _, err := ParseToken("secret-b", token); err == nil {
		t.Error("expected an error when parsing a token with the wrong secret")
	}
}

func TestParseToken_ExpiredRejected(t *testing.T) {
	token, _ := GenerateToken("secret", -time.Hour, 1, models.RoleCustomer)
	if _, err := ParseToken("secret", token); err == nil {
		t.Error("expected an error when parsing an expired token")
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Go Conf 2024!":       "go-conf-2024",
		"  Leading/Trailing ": "leading-trailing",
		"Múltiple   Spaces":   "m-ltiple-spaces",
	}
	for input, want := range cases {
		if got := Slugify(input); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRandomCode_LengthAndUniqueness(t *testing.T) {
	a := RandomCode(8)
	b := RandomCode(8)
	if len(a) != 8 || len(b) != 8 {
		t.Fatalf("expected length 8, got %d and %d", len(a), len(b))
	}
	if a == b {
		t.Error("two consecutive random codes should not collide")
	}
}
