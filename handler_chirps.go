package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	p := params{}
	err := decoder.Decode(&p)
	if err != nil {
		respondWithError(w, 400, "Couldn't decode params", err)
		return
	}
	if len(p.Body) > 140 {
		log.Printf("Chirp too long")
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
	respondWithJSON(w, 200, struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanedBody,
	})
}
