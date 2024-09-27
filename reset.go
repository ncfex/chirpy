package main

import "net/http"

func (api *apiConfig) handlerReset(rw http.ResponseWriter, r *http.Request) {
	if api.platform != "DEV" {
		respondWithError(rw, http.StatusForbidden, "forbidden")
		return
	}

	err := api.DB.Reset(r.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "internal server error")
		return
	}
	api.fileserverHits = 0
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("reset hits and db"))
}
