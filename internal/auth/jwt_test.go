package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  http.Header
		want    string
		wantErr bool
	}{
		{"valid", http.Header{"Authorization": {"Bearer abc.def.ghi"}}, "abc.def.ghi", false},
		{"missing header", http.Header{}, "", true},
		{"empty header", http.Header{"Authorization": {""}}, "", true},
		{"wrong scheme", http.Header{"Authorization": {"Basic abc"}}, "", true},
		{"lowercase scheme", http.Header{"Authorization": {"bearer abc"}}, "", true},
		{"no token", http.Header{"Authorization": {"Bearer"}}, "", true},
		{"extra parts", http.Header{"Authorization": {"Bearer abc def"}}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBearerToken(tt.header)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
