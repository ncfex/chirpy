package main

import (
	"encoding/json"
	"net/http"

	"github.com/ncfex/chirpy/internal/auth"
	"github.com/ncfex/chirpy/internal/database"
)

func (cfg *apiConfig) HandlerPolkaWebhook(rw http.ResponseWriter, r *http.Request) {
	polkaAPIKey, err := auth.GetAuthorizationHeaderItem(&r.Header, "ApiKey")
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "no permission")
		return
	}

	if cfg.polkaAPIKey != polkaAPIKey {
		respondWithError(rw, http.StatusUnauthorized, "no permission")
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(rw, http.StatusNoContent, struct{}{})
		return
	}

	err = cfg.DB.UpgradeUserToChirpyRed(params.Data.UserID)
	if err != nil {
		statusCode := http.StatusNoContent
		if err == database.ErrNotExist {
			statusCode = http.StatusNotFound
		}
		respondWithError(rw, statusCode, err.Error())
		return
	}

	respondWithJSON(rw, http.StatusNoContent, struct{}{})
}
