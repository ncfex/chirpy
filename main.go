package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
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
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
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

func handlerChirpsValidate(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Body string `json:"body"`
	}

	type validResponseParams struct {
		Valid bool `json:"valid"`
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

	respondWithJSON(rw, http.StatusOK, validResponseParams{
		Valid: true,
	})
}
