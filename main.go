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
	platform       string
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
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	if len(jwtSecret) == 0 || len(polkaAPIKey) == 0 || len(dbURL) == 0 || len(platform) == 0 {
		log.Fatal("INVALID env VALUES")
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
		platform:       platform,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/users", apiCfg.handlerUserGetAll)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUserCreate)

	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.handlerChirpGetById)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpGetAll)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpCreate)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
