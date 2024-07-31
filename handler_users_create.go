package main

import (
	"encoding/json"
	"net/http"
)

type User struct {
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerNewUser(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	if params.Email == "" {
		respondWithError(rw, http.StatusBadRequest, "Invalid email format")
		return
	}

	addedUser, err := cfg.DB.CreateUser(params.Email)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error creating new user")
		return
	}

	respondWithJSON(rw, http.StatusCreated, addedUser)
}
