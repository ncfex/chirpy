package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ncfex/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
}

const DATABASE_FILE_NAME = "database.json"

func main() {
	const filepathRoot = "."
	const port = "8080"

	db, err := database.NewDb(DATABASE_FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", readiness)
	mux.HandleFunc("GET /api/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerNewChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handlerGetChirpById)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handlerNewChirp(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(rw, http.StatusBadRequest, "Chirp is too long")
		return
	}

	bannedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleaned_s := handleBannedWords(bannedWords, params.Body)
	newC, createChipErr := cfg.DB.CreateChirp(cleaned_s)
	if createChipErr != nil {
		log.Printf("Error creating new chip: %s", createChipErr)
		respondWithError(rw, http.StatusInternalServerError, "Error creating new chip")
		return
	}

	respondWithJSON(rw, http.StatusCreated, newC)
}

func (cfg *apiConfig) handlerGetChirps(rw http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		respondWithError(rw, http.StatusInternalServerError, "Error getting chirps")
		return
	}

	respondWithJSON(rw, http.StatusOK, chirps)
}

func handleBannedWords(bannedWords map[string]struct{}, s string) string {
	splitted := strings.Split(s, " ")
	for i, word := range splitted {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			splitted[i] = "****"
		}
	}
	return strings.Join(splitted, " ")
}

func (cfg *apiConfig) handlerGetChirpById(rw http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		log.Printf("error getting  chirps")
		respondWithError(rw, 500, "error getting  chirps")
		return
	}

	chirpId, toIntErr := strconv.Atoi(r.PathValue("chirpId"))
	if toIntErr != nil {
		log.Printf("not valid ID")
		respondWithError(rw, 500, "not valid ID")
		return
	}

	chirpToReturn := database.Chirp{}
	for _, chirp := range chirps {
		if chirp.Id == chirpId {
			chirpToReturn = chirp
		}
	}

	if chirpToReturn.Id == 0 {
		log.Printf("chirp not found")
		respondWithError(rw, 404, "chirp not found")
		return
	}

	respondWithJSON(rw, 200, chirpToReturn)
}
