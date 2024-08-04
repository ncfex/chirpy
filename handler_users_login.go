package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ncfex/chirpy/internal/auth"
)

func (cfg *apiConfig) HandlerLogin(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		User
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	user, err := cfg.DB.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't get user")
		return
	}

	err = auth.CheckPasswordHash(params.Password, string(user.Password))
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Invalid password")
		return
	}

	defaultExpiration := 60 * 60 * 24
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = defaultExpiration
	} else if params.ExpiresInSeconds > defaultExpiration {
		params.ExpiresInSeconds = defaultExpiration
	}

	token, err := auth.GenerateJWT("chirpy", string(cfg.jwtSecret), auth.UserJWTPayload{
		Id: user.Id,
	}, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJSON(rw, http.StatusOK, response{
		User: User{
			Id:    user.Id,
			Email: user.Email,
		},
		Token: token,
	})
}
