package main

import (
	"net/http"
	"strconv"

	"github.com/ncfex/chirpy/internal/auth"
)

func (cfg *apiConfig) HandlerDeleteChirpById(rw http.ResponseWriter, r *http.Request) {
	chirpId, err := strconv.Atoi(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(rw, http.StatusBadRequest, "invalid chirp id")
		return
	}

	tokenString, err := auth.GetBearerToken(&r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "no permission")
		return
	}

	userIDString, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	userIDInt, err := strconv.Atoi(userIDString)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	err = cfg.DB.DeleteChirp(chirpId, userIDInt)
	if err != nil {
		respondWithError(rw, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusNoContent, struct{}{})
}
