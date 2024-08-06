package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ncfex/chirpy/internal/auth"
	"github.com/ncfex/chirpy/internal/database"
)

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"-"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerNewUser(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	if params.Email == "" || params.Password == "" {
		respondWithError(rw, http.StatusBadRequest, "Invalid email format")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	addedUser, err := cfg.DB.CreateUser(params.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyExists) {
			respondWithError(rw, http.StatusConflict, "User already exists")
			return
		}

		respondWithError(rw, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJSON(rw, http.StatusCreated, response{
		User: User{
			Id:    addedUser.Id,
			Email: addedUser.Email,
		},
	})
}
