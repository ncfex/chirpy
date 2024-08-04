package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) HandlerUserUpdate(rw http.ResponseWriter, r *http.Request) {
	token := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
	if token == "" {
		respondWithError(rw, http.StatusUnauthorized, "Please provide token")
		return
	}

	decoded, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, err.Error())
		return
	}

	type reqBodyParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := reqBodyParams{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error while decoding")
		return
	}

	userId, err := decoded.Claims.GetSubject()
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, err.Error())
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "error while converting")
		return
	}

	updatedUser, err := cfg.DB.UpdateUser(userIdInt, struct {
		Email    string
		Password string
	}{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		respondWithError(rw, http.StatusUnauthorized, "Error updating user")
		return
	}

	respondWithJSON(rw, http.StatusOK, struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}{
		Id:    updatedUser.Id,
		Email: updatedUser.Email,
	})
}
