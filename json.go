package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("error marshalling json: ", payload)
		w.WriteHeader(500)
		return
	}

	// Set the content type to JSON
	w.Header().Add("Content-Type", "application/json")
	// Set the status code
	w.WriteHeader(code)
	// Encode the payload to JSON
	w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	if code > 499 {
		log.Println("Responding with 5XX level error: ", message)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, code, errorResponse{Error: message})
}
