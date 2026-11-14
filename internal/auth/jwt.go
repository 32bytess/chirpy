package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	})
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString,
		&claimsStruct,
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("chirpy-access"))
	if err != nil {
		return uuid.UUID{}, err
	}
	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := headers.Get("Authorization")
	if len(tokenString) == 0 {
		return "", fmt.Errorf("no token")
	}
	words := strings.Split(tokenString, " ")
	if len(words) != 2 {
		return "", fmt.Errorf("unvalid token")
	}
	if words[0] != "Bearer" {
		return "", fmt.Errorf("unvalid token")
	}
	return words[1], nil
}
