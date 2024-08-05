package main

import (
	"net/http"
	"time"

	"github.com/ncfex/chirpy/internal/auth"
)

func (cfg *apiConfig) HandlerRefreshToken(rw http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(&r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Couldn't find Refresh Token")
		return
	}

	user, err := cfg.DB.GetUserByRefreshToken(refreshToken)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Invalid Refresh Token")
		return
	}

	if user.RefreshToken.Exp < time.Now().UTC().Unix() {
		_, err := cfg.DB.LogoutUser(user.Id)
		if err != nil {
			respondWithError(rw, http.StatusUnauthorized, "Error while logout")
			return
		}
		respondWithError(rw, http.StatusUnauthorized, "Refresh Token is Expired")
		return
	}

	token, err := auth.GenerateJWT("chirpy", string(cfg.jwtSecret), auth.UserJWTPayload{
		Id: user.Id,
	}, time.Duration(60*60)*time.Second)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJSON(rw, http.StatusOK, response{
		Token: token,
	})
}

func (cfg *apiConfig) HandlerRevokeRefresh(rw http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(&r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Couldn't find Refresh Token")
		return
	}

	user, err := cfg.DB.GetUserByRefreshToken(refreshToken)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Invalid Refresh Token")
		return
	}

	_, err = cfg.DB.LogoutUser(user.Id)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while logout")
		return
	}

	respondWithJSON(rw, http.StatusNoContent, struct{}{})
}
