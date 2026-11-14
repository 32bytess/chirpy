package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	id := uuid.New()

	token, err := MakeJWT(id, "secret", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT: %v", err)
	}

	got, err := ValidateJWT(token, "secret")
	if err != nil {
		t.Fatalf("ValidateJWT: %v", err)
	}
	if got != id {
		t.Errorf("got %v, want %v", got, id)
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	token, _ := MakeJWT(uuid.New(), "secret", time.Hour)

	if _, err := ValidateJWT(token, "wrong"); err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestValidateJWTExpired(t *testing.T) {
	token, _ := MakeJWT(uuid.New(), "secret", -time.Hour)

	if _, err := ValidateJWT(token, "secret"); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateJWTMalformed(t *testing.T) {
	if _, err := ValidateJWT("not-a-jwt", "secret"); err == nil {
		t.Error("expected error for malformed token")
	}
}
