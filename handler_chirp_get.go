package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerChirpGetAll(rw http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.DB.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	chirpMap := make([]Chirp, 0, len(chirps))
	for _, c := range chirps {
		chirpMap = append(chirpMap, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}

	respondWithJSON(rw, http.StatusOK, chirpMap)
}
