package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email string `json:"email"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 400, "Couldn't decode params", err)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
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
