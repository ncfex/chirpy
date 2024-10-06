package main

import (
	"net/http"

	"github.com/google/uuid"
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

func (cfg *apiConfig) handlerChirpGetById(rw http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("id")
	if chirpID == "" {
		respondWithError(rw, http.StatusNotFound, "not found")
		return
	}

	pchirpID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(rw, http.StatusNotFound, "invalid ID")
		return
	}

	c, err := cfg.DB.GetChirpById(req.Context(), pchirpID)
	if err != nil {
		respondWithError(rw, http.StatusNotFound, err.Error())
		return
	}

	mC := Chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}

	respondWithJSON(rw, http.StatusOK, mC)
}
