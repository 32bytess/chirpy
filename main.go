package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	handlerHome := http.StripPrefix("/app", http.FileServer(http.Dir("app/")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(handlerHome))
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics())
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	log.Fatal(srv.ListenAndServe())
}
