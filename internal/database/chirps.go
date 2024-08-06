package database

import "errors"

var ErrNoPermission = errors.New("no permission")
var ErrChirpNotFound = errors.New("chirp not found")

type Chirp struct {
	Id       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	newChirp := Chirp{
		Id:       id,
		Body:     body,
		AuthorID: authorId,
	}
	dbStructure.Chirps[id] = newChirp

	if writeErr := db.writeDB(dbStructure); writeErr != nil {
		return Chirp{}, nil
	}
	return newChirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, v := range dbStructure.Chirps {
		chirps = append(chirps, v)
	}

	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, ErrChirpNotFound
	}

	return chirp, nil
}

func (db *DB) DeleteChirp(chirpID int, requesterID int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	chirp, ok := dbStructure.Chirps[chirpID]
	if !ok {
		return ErrChirpNotFound
	}

	if chirp.AuthorID != requesterID {
		return ErrNoPermission
	}

	delete(dbStructure.Chirps, chirpID)

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}
