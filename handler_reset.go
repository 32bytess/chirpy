package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		respondWithError(w, 403, "403 Forbidden", nil)
		return
	}
	cfg.fileserverHits.Swap(0)
	err := cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	}
	respondWithJSON(w, 200, struct {
		Message string `json:"message"`
	}{
		Message: "Hits reset to 0 and delete all users",
	})
}
