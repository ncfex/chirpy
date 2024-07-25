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
	apiCfg := apiConfig{}

	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", readiness)
	mux.HandleFunc("GET /api/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (api *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		api.fileserverHits++
		next.ServeHTTP(rw, r)
	})
}

func validateChirpHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	type bodyParams struct {
		Body string `json:"body"`
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	type validResponse struct {
		Valid bool `json:"valid"`
	}

	somethingWentWrongErr := errorResponse{
		Error: "Something went wrong",
	}
	errDat, smthWentWErr := json.Marshal(somethingWentWrongErr)
	if smthWentWErr != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(errDat)
		return
	}

	chirpTooLongErr := errorResponse{
		Error: "Chirp is too long",
	}
	tooLongErrDat, toLongErr := json.Marshal(chirpTooLongErr)
	if toLongErr != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(tooLongErrDat)
		return
	}

	validResp := validResponse{
		Valid: true,
	}
	validRespDat, validRespErr := json.Marshal(validResp)
	if validRespErr != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(errDat)
		return
	}

	// parse r body
	decoder := json.NewDecoder(r.Body)
	params := bodyParams{}

	decodeErr := decoder.Decode(&params)
	if decodeErr != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(errDat)
		return
	}

	if len(params.Body) > 140 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(tooLongErrDat)
		return
	} else {
		rw.WriteHeader(http.StatusOK)
		rw.Write(validRespDat)
		return
	}
}
