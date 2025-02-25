package main

import "net/http"

func getRouter(apiCfg *apiConfig) *http.ServeMux {
	mux := http.NewServeMux()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(apiCfg.filepathRoot))))

	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handleUpgradeUser)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleAuthenticateUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevokeRefreshToken)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	return mux
}
