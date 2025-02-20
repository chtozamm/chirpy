package main

import (
	"context"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	err := cfg.db.RemoveAllUsers(context.Background())
	if err != nil {
		log.Printf("Error removing all users: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
