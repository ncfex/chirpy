package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerGetChirps(rw http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	authorID := -1
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString != "" {
		authorID, err = strconv.Atoi(authorIDString)
		if err != nil {
			respondWithError(rw, http.StatusBadRequest, "Invalid author ID")
			return
		}
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		if authorID != -1 && dbChirp.AuthorID != authorID {
			continue
		}

		chirps = append(chirps, Chirp{
			ID:       dbChirp.Id,
			AuthorID: dbChirp.AuthorID,
			Body:     dbChirp.Body,
		})
	}

	sortType := "asc"
	sortParam := r.URL.Query().Get("sort")
	if sortParam != "" {
		sortType = sortParam
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortType == "asc" {
			return chirps[i].ID < chirps[j].ID
		}
		return chirps[i].ID > chirps[j].ID
	})

	respondWithJSON(rw, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirpById(rw http.ResponseWriter, r *http.Request) {
	chirpId, toIntErr := strconv.Atoi(r.PathValue("chirpId"))
	if toIntErr != nil {
		respondWithError(rw, http.StatusInternalServerError, "not valid ID")
		return
	}

	chirp, err := cfg.DB.GetChirp(chirpId)
	if err != nil {
		respondWithError(rw, http.StatusInternalServerError, "Error getting chirp")
		return
	}

	respondWithJSON(rw, http.StatusOK, chirp)
}
