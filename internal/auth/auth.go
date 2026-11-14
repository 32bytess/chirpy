package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("something went wrong : %w", err)
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	check, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("something went wrong : %w", err)
	}
	return check, nil
}
