package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type resOK struct {
		Valid bool `json:"valid"`
	}

	type resErr struct {
		Error string `json:"error"`
	}

	if len(params.Body) > 140 {
		response, err := json.Marshal(resErr{Error: "Chirp is too long"})
		if err != nil {
			log.Printf("Error marshalling JSON: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	response, err := json.Marshal(resOK{Valid: true})
	if err != nil {
		log.Printf("Error marshalling JSON: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handlerCensor(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type resOK struct {
		CleanedBody string `json:"cleaned_body"`
	}

	profanities := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	var cleaned string
	for _, profanity := range profanities {
		cleaned = strings.ReplaceAll(params.Body, profanity, "****")
	}

	response, err := json.Marshal(resOK{CleanedBody: cleaned})
	if err != nil {
		log.Printf("Error marshalling JSON: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
