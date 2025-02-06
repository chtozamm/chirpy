package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// mux.HandleFunc("/", http.NotFound)

	log.Fatal(server.ListenAndServe())
}
