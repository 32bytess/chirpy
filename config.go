package main

import (
	"sync/atomic"

	"github.com/32bytess/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
}
