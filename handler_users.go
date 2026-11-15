package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/32bytess/chirpy/internal/auth"
	"github.com/32bytess/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Couldn't decode params", err)
		return
	}

	if len(params.Email) == 0 || len(params.Password) == 0 {
		respondWithError(w, http.StatusBadRequest, "Bad Request", nil)
		return
	}
	// need some validation for the password later (strength)

	// hash pass
	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:    params.Email,
		Password: hashedPass,
	})
	if err != nil {
		respondWithError(w, 400, "User didnt get created", err)
		return
	}
	respondWithJSON(w, 201, User{
		ID:        uuid.UUID(user.ID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Couldn't decode params", err)
		return
	}

	if len(params.Email) == 0 || len(params.Password) == 0 {
		respondWithError(w, http.StatusBadRequest, "Unvalid inputs", nil)
		return
	}
	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}
	// compare passwords
	valid, err := auth.CheckPasswordHash(params.Password, user.Password)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	if !valid {
		respondWithError(w, http.StatusUnauthorized, "Unvalid Password", nil)
		return
	}
	id := uuid.UUID(user.ID)
	token, err := auth.MakeJWT(id, cfg.TokenSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    id,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	respondWithJSON(w, 200, User{
		ID:           id,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unvalid token", err)
		return
	}
	refreshToken, err := cfg.DB.GetRefreshTokenByToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
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
