package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/32bytess/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	p := params{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		respondWithError(w, 400, "Couldn't decode params", err)
		return
	}
	if len(p.Body) == 0 {
		respondWithError(w, 400, "body can't be nil", err)
		return
	}

	if len(p.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	words := strings.Fields(p.Body)

	for i := range words {
		lower := strings.ToLower(words[i])
		if _, ok := badWords[lower]; ok {
			words[i] = "****"
		}
	}
	cleanedBody := strings.Join(words, " ")
	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		respondWithError(w, 400, "Unvalid user id", err)
		return
	}
	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, 400, "Chirp didnt get created", err)
		return
	}
	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
