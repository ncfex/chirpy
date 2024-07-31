package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/ncfex/chirpy/internal/database"
)

func (cfg *apiConfig) handlerGetChirps(rw http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		respondWithError(rw, http.StatusInternalServerError, "Error getting chirps")
		return
	}

	respondWithJSON(rw, http.StatusOK, chirps)
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
