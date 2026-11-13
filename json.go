package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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
