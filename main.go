package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	mux.HandleFunc("/healthz", readiness)
	mux.HandleFunc("/metrics", apiCfg.getMetrics)
	mux.HandleFunc("/reset", apiCfg.resetMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func readiness(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	_, err := rw.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		os.Exit(1)
	}
}

func (api *apiConfig) getMetrics(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	_, err := rw.Write([]byte(fmt.Sprintf("Hits: %v", api.fileserverHits)))
	if err != nil {
		fmt.Println("can't write to body")
	}
}

func (api *apiConfig) resetMetrics(rw http.ResponseWriter, r *http.Request) {
	api.fileserverHits = 0
	_, err := rw.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		fmt.Println("can't write to body")
	}
}

func (api *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		api.fileserverHits++
		next.ServeHTTP(rw, r)
	})
}
