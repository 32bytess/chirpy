package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == password {
		t.Error("hash should not equal the plaintext password")
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected argon2id hash, got %q", hash)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == hash2 {
		t.Error("hashing the same password twice should give different hashes (random salt)")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("correctPassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{"correct password", "correctPassword", hash, true, false},
		{"wrong password", "wrongPassword", hash, false, false},
		{"empty password", "", hash, false, false},
		{"invalid hash", "correctPassword", "not-a-hash", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
