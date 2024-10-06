package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

func respError(w http.ResponseWriter, statusCode int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}

	if statusCode > 499 {
		log.Printf("Responding with 5XX error: %s\n", msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	respJSON(w, statusCode, errorResponse{
		Error: msg,
	})
}
