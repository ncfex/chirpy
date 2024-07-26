package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDb(path string) (*DB, error) {
	rwMutex := sync.RWMutex{}
	db := &DB{
		path: path,
		mux:  &rwMutex,
	}

	if err := db.ensureDB(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, loadDbErr := db.loadDB()
	if loadDbErr != nil {
		log.Printf("Error loading db: %s", loadDbErr)
		return Chirp{}, loadDbErr
	}

	var newID int
	for id := range dbStructure.Chirps {
		if id >= newID {
			newID = id + 1
		}
	}
	if len(dbStructure.Chirps) == 0 {
		newID = 1
	}

	newChirp := Chirp{
		Id:   newID,
		Body: body,
	}

	dbStructure.Chirps[newChirp.Id] = newChirp

	if writeErr := db.writeDB(dbStructure); writeErr != nil {
		log.Printf("Error on writeDB function: %s", writeErr)
		return Chirp{}, nil
	}
	return newChirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, loadDbErr := db.loadDB()
	if loadDbErr != nil {
		log.Printf("Error loading db: %s", loadDbErr)
		return nil, loadDbErr
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, v := range dbStructure.Chirps {
		chirps = append(chirps, v)
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})

	return chirps, nil
}

func (db *DB) ensureDB() error {
	_, fileReadErr := os.ReadFile(db.path)
	if fileReadErr != nil {
		if errors.Is(fileReadErr, os.ErrNotExist) {
			dbStructure := DBStructure{
				Chirps: make(map[int]Chirp),
			}
			dbData, err := json.Marshal(dbStructure)
			if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				return err
			}
			writeErr := os.WriteFile(db.path, dbData, 0666)
			if writeErr != nil {
				log.Printf("Error writing to file: %s", writeErr)
				return writeErr
			}
			return nil
		}
		log.Printf("Error reading file %s", fileReadErr)
		return fileReadErr
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	err := db.ensureDB()
	if err != nil {
		return DBStructure{}, err
	}

	fileData, fileReadErr := os.ReadFile(db.path)
	if fileReadErr != nil {
		log.Printf("Error reading file: %s", fileReadErr)
		return DBStructure{}, fileReadErr
	}

	var dbStructure DBStructure
	unMarshalErr := json.Unmarshal(fileData, &dbStructure)
	if unMarshalErr != nil {
		log.Printf("Error unmarshalling JSON: %s", unMarshalErr)
		return DBStructure{}, unMarshalErr
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	writeData, marshalErr := json.Marshal(dbStructure)
	if marshalErr != nil {
		log.Printf("Error while marshaling %s", marshalErr)
		return marshalErr
	}

	if writeErr := os.WriteFile(db.path, writeData, 0666); writeErr != nil {
		log.Printf("Error writing to db file: %s", writeErr)
		return writeErr
	}

	return nil
}
