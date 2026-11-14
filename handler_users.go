package main

import (
	"encoding/json"
	"net/http"

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
	respondWithJSON(w, 200, User{
		ID:        uuid.UUID(user.ID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
