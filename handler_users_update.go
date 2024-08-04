package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ncfex/chirpy/internal/auth"
)

func (cfg *apiConfig) HandlerUserUpdate(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
	}

	token, err := auth.GetBearerToken(&r.Header)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	userIDInt, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	updatedUser, err := cfg.DB.UpdateUser(userIDInt, params.Email, hashedPassword)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Error updating user")
		return
	}

	respondWithJSON(rw, http.StatusOK, response{
		User: User{
			Id:    updatedUser.Id,
			Email: updatedUser.Email,
		},
	})
}
