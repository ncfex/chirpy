package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/ncfex/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.Queries
	jwtSecret      string
	polkaAPIKey    string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	polkaAPIKey := os.Getenv("POLKA_API_KEY")
	if len(jwtSecret) == 0 || len(polkaAPIKey) == 0 {
		log.Fatal("Secret key error")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             dbQueries,
		jwtSecret:      jwtSecret,
		polkaAPIKey:    polkaAPIKey,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", readiness)
	mux.HandleFunc("GET /api/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerNewChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handlerGetChirpById)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", apiCfg.HandlerDeleteChirpById)

	mux.HandleFunc("POST /api/users", apiCfg.handlerNewUser)
	mux.HandleFunc("PUT /api/users", apiCfg.HandlerUserUpdate)
	mux.HandleFunc("POST /api/login", apiCfg.HandlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.HandlerRevokeRefresh)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandlerPolkaWebhook)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
