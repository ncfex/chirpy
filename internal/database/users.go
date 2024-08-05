package database

import (
	"errors"
	"time"
)

var ErrAlreadyExists = errors.New("already exists")

type RefreshToken struct {
	Token string `json:"refresh_token"`
	Exp   int64  `json:"exp"`
}

type User struct {
	Id           int          `json:"id"`
	Email        string       `json:"email"`
	Password     []byte       `json:"password"`
	RefreshToken RefreshToken `json:"refresh_token"`
}

func (db *DB) CreateUser(email string, hashedPassword []byte) (User, error) {
	if _, err := db.GetUserByEmail(email); !errors.Is(err, ErrNotExist) {
		return User{}, ErrAlreadyExists
	}

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1
	newUser := User{
		Id:       id,
		Email:    email,
		Password: hashedPassword,
	}

	dbStructure.Users[id] = newUser

	if err = db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (db *DB) GetUsers() ([]User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(dbStructure.Users))
	for _, user := range dbStructure.Users {
		users = append(users, user)
	}

	return users, nil
}

func (db *DB) GetUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotExist
}

func (db *DB) GetUserByRefreshToken(refreshTokenString string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.RefreshToken.Token == refreshTokenString {
			return user, nil
		}
	}

	return User{}, ErrNotExist
}

func (db *DB) UpdateUser(id int, email string, hashedPassword []byte) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	user.Password = hashedPassword
	user.Email = email
	dbStructure.Users[id] = user

	if err = db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) LoginUser(id int, refreshTokenStr string, refreshTokenDuration time.Duration) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	var refreshToken RefreshToken
	if refreshTokenStr != "" || refreshTokenDuration == 0 {
		refreshToken = RefreshToken{
			Token: refreshTokenStr,
			Exp:   time.Now().UTC().Add(refreshTokenDuration).Unix(),
		}
	}
	var zeroToken RefreshToken
	if zeroToken != refreshToken {
		user.RefreshToken = refreshToken
	}

	dbStructure.Users[id] = user

	if err = db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) LogoutUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	user.RefreshToken = RefreshToken{}
	dbStructure.Users[id] = user

	if err = db.writeDB(dbStructure); err != nil {
		return User{}, err
	}

	return user, nil
}
