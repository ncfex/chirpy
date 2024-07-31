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
}

var db *database.DB

const DATABASE_FILE_NAME = "database.json"

func main() {
	var dbErr error
	db, dbErr = database.NewDb(DATABASE_FILE_NAME)
	if dbErr != nil {
		log.Fatalf("Failed to initialize the database: %v", dbErr)
		return
	}

	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/*", fsHandler)

	mux.HandleFunc("GET /api/healthz", readiness)
	mux.HandleFunc("GET /api/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/chirps", handlerNewChirp)
	mux.HandleFunc("GET /api/chirps", handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handlerGetChirpById)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func parseJSONBody[T any](decoder *json.Decoder, v *T) error { // Test
	return decoder.Decode(&v)
}

func respondWithJSON(rw http.ResponseWriter, code int, payload interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		rw.WriteHeader(500)
		return
	}
	rw.WriteHeader(code)
	rw.Write(dat)
}

func respondWithError(rw http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(rw, code, errorResponse{
		Error: msg,
	})
}

func handlerNewChirp(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err := parseJSONBody(decoder, &params)
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
	newC, createChipErr := db.CreateChirp(cleaned_s)
	if createChipErr != nil {
		log.Printf("Error creating new chip: %s", createChipErr)
		respondWithError(rw, http.StatusInternalServerError, "Error creating new chip")
		return
	}

	respondWithJSON(rw, http.StatusCreated, newC)
}

func handlerGetChirps(rw http.ResponseWriter, r *http.Request) {
	chirps, err := db.GetChirps()
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

func (api *apiConfig) handlerGetChirpById(rw http.ResponseWriter, r *http.Request) {
	chirps, err := db.GetChirps()
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
