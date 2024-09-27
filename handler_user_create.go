package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUserCreate(rw http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Internal server error")
		return
	}

	user, err := cfg.DB.CreateUser(req.Context(), params.Email)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error creating user")
		return
	}

	respondWithJSON(rw, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}
