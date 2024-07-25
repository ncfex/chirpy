package main

import "net/http"

func (api *apiConfig) resetMetrics(rw http.ResponseWriter, r *http.Request) {
	api.fileserverHits = 0
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Hits reset to 0"))
}
