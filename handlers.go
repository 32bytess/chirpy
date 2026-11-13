package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	p := params{}
	err := decoder.Decode(&p)
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
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

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	html := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	val := fmt.Sprintf(html, cfg.fileserverHits.Load())
	w.Write([]byte(val))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Swap(0)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email string `json:"email"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 400, "User didnt get created", err)
		return
	}
	respondWithJSON(w, 201, user)
}

// helpers

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		log.Printf(" err: %s", err)
	}
	type ErrorMessage struct {
		Message string `json:"error"`
	}
	w.WriteHeader(code)
	errorMessage := ErrorMessage{
		Message: msg,
	}
	res, _ := json.Marshal(errorMessage)
	w.Write(res)
}

func respondWithJSON(w http.ResponseWriter, code int, payloud any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	res, _ := json.Marshal(payloud)
	w.Write(res)
}
