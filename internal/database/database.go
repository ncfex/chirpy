package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
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

func (db *DB) GetUsers() ([]User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		log.Printf("Error loading db: %s", err)
		return nil, err
	}

	users := make([]User, 0, len(dbStructure.Users))
	for _, user := range dbStructure.Users {
		users = append(users, user)
	}

	return users, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	var newID int
	for id := range dbStructure.Users {
		if id >= newID {
			newID = id + 1
		}
	}
	if len(dbStructure.Users) == 0 {
		newID = 1
	}

	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return User{}, err
	}

	newUser := User{
		Id:       newID,
		Email:    email,
		Password: hashedPw,
	}

	dbStructure.Users[newID] = newUser

	if err = db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (db *DB) UpdateUser(id int, payload struct {
	Email    string
	Password string
}) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, errors.New("User not found")
	}

	newHashedPw, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 10)
	if err != nil {
		return User{}, err
	}

	user.Password = newHashedPw
	user.Email = payload.Email
	dbStructure.Users[id] = user

	if err = db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) RemoveDBFileForDebug() error {
	err := os.Remove(db.path)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) ensureDB() error {
	_, fileReadErr := os.ReadFile(db.path)
	if fileReadErr != nil {
		if errors.Is(fileReadErr, os.ErrNotExist) {
			dbStructure := DBStructure{
				Chirps: make(map[int]Chirp),
				Users:  make(map[int]User),
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
