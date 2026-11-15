package main

import (
	"net/http"
	"time"

	"github.com/32bytess/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}
	refreshToken, err := cfg.DB.GetRefreshTokenByToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	if refreshToken.ExpiresAt.Before(time.Now()) || refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.TokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unvalid token", err)
		return
	}
	if err = cfg.DB.RevokeToken(r.Context(), token); err != nil {
		respondWithError(w, http.StatusBadRequest, "Unvalid token", err)
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}
