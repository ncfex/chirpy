package main

import (
	"crypto/rand"
	"encoding/hex"
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
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.GenerateJWT("chirpy", string(cfg.jwtSecret), auth.UserJWTPayload{
		Id: user.Id,
	}, time.Duration(60*60)*time.Second)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	random := make([]byte, 32)
	_, err = rand.Read(random)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error generating random")
		return
	}

	refreshTokenStr := hex.EncodeToString(random)
	refreshTokenDuration := time.Duration(24*60) * time.Hour

	_, err = cfg.DB.LoginUser(user.Id, refreshTokenStr, refreshTokenDuration)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error on login")
		return
	}

	respondWithJSON(rw, http.StatusOK, response{
		User: User{
			Id:    user.Id,
			Email: user.Email,
		},
		Token:        token,
		RefreshToken: refreshTokenStr,
	})
}
