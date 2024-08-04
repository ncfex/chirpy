package main

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) HandlerLogin(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	users, err := cfg.DB.GetUsers()
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while getting users")
		return
	}

	if len(users) == 0 {
		respondWithError(rw, http.StatusNotFound, "User not found")
		return
	}

	for _, user := range users {
		if user.Email == params.Email {
			err := bcrypt.CompareHashAndPassword(user.Password, []byte(params.Password))
			if err != nil {
				respondWithError(rw, http.StatusUnauthorized, "Incorrect password")
				return
			}

			expInS := 0
			if params.ExpiresInSeconds != 0 {
				expInS = params.ExpiresInSeconds
			}

			duration := time.Second * time.Duration(expInS)
			token, err := cfg.GenerateJWT("chirpy", UserJWTPayload{
				Id: user.Id,
			}, duration)
			if err != nil {
				respondWithError(rw, http.StatusInternalServerError, err.Error())
				return
			}

			respondWithJSON(rw, http.StatusOK, struct {
				Id    int    `json:"id"`
				Email string `json:"email"`
				Token string `json:"token"`
			}{
				Id:    user.Id,
				Email: user.Email,
				Token: token,
			})
			return
		}
	}

	respondWithError(rw, http.StatusBadRequest, "Invalid credentials")
}
