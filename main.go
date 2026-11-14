package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/32bytess/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db err: %v\n", err)
	}

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		DB:             database.New(db),
		Platform:       os.Getenv("PLATFORM"),
	}
	mux := http.NewServeMux()

	// app
	handlerHome := http.StripPrefix("/app", http.FileServer(http.Dir("app/")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(handlerHome))

	// admin
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	// api
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirp)
	port := os.Getenv("PORT")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
