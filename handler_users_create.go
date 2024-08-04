package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerNewUser(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	users, err := cfg.DB.GetUsers()
	if err != nil {
		respondWithError(rw, http.StatusBadRequest, "Error while getting users")
		return
	}

	for _, user := range users {
		if user.Email == params.Email {
			respondWithError(rw, http.StatusBadRequest, "Email already exists")
			return
		}
	}

	addedUser, err := cfg.DB.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error creating new user")
		return
	}

	respondWithJSON(rw, http.StatusCreated, struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}{
		Id:    addedUser.Id,
		Email: addedUser.Email,
	})
}
