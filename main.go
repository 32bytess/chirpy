package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/32bytess/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
}

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("db err: %v\n", err)
		os.Exit(1)
	}

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		DB:             database.New(db),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	handlerHome := http.StripPrefix("/app", http.FileServer(http.Dir("app/")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(handlerHome))
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /api/user", cfg.handlerCreateUser)
	port := os.Getenv("PORT")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	log.Fatal(srv.ListenAndServe())
}
