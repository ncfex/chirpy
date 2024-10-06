package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerUserGetAll(rw http.ResponseWriter, req *http.Request) {
	users, err := cfg.DB.GetAllUsers(req.Context())
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	mainUsers := make([]User, 0, len(users))
	for _, u := range users {
		mainUsers = append(mainUsers, User{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Email:     u.Email,
		})
	}

	respondWithJSON(rw, http.StatusOK, mainUsers)
}
