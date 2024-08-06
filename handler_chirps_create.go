package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/ncfex/chirpy/internal/auth"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (cfg *apiConfig) handlerNewChirp(rw http.ResponseWriter, r *http.Request) {
	type reqBodyParams struct {
		Body string `json:"body"`
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

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, err.Error())
		return
	}

	userIDInt, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}
	newC, createChipErr := cfg.DB.CreateChirp(cleaned, userIDInt)
	if createChipErr != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error creating new chip")
		return
	}

	respondWithJSON(rw, http.StatusCreated, newC)
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("chirp is too long")
	}

	bannedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	return handleBannedWords(bannedWords, body), nil
}

func handleBannedWords(bannedWords map[string]struct{}, s string) string {
	splitted := strings.Split(s, " ")
	for i, word := range splitted {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			splitted[i] = "****"
		}
	}
	return strings.Join(splitted, " ")
}
