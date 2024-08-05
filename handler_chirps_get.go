package main

import (
	"net/http"
	"strconv"
)

func (cfg *apiConfig) handlerGetChirps(rw http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error getting chirps")
		return
	}

	respondWithJSON(rw, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirpById(rw http.ResponseWriter, r *http.Request) {
	chirpId, toIntErr := strconv.Atoi(r.PathValue("chirpId"))
	if toIntErr != nil {
		respondWithError(rw, http.StatusInternalServerError, "not valid ID")
		return
	}

	chirp, err := cfg.DB.GetChirp(chirpId)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error getting chirp")
		return
	}

	respondWithJSON(rw, http.StatusOK, chirp)
}
