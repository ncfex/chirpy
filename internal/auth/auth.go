package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserJWTPayload struct {
	Id int
}

const DEFAULT_EXPIRES_DURATION = time.Hour * 24

func HashPassword(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return []byte(""), err
	}

	return hashedPassword, nil
}

func CheckPasswordHash(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func GenerateJWT(issuer string, tokenSecret string, payload UserJWTPayload, expiresInSeconds time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresInSeconds)),
		Subject:   fmt.Sprintf("%d", payload.Id),
	})

	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString string, tokenSecret string) (string, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(t *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}

	if claimsStruct.ExpiresAt.Before(time.Now()) {
		return "", errors.New("expired")
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}
	if issuer != string("chirpy") {
		return "", errors.New("invalid issuer")
	}

	return userIDString, nil
}

func GetBearerToken(headers *http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no auth header")
	}

	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}

	return splitAuth[1], nil
}

func GetAuthorizationHeaderItem(headers *http.Header, itemKeyString string) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no auth header")
	}

	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != itemKeyString {
		return "", errors.New("invalid auth header")
	}

	return splitAuth[1], nil
}

func GenerateRefreshToken() (string, time.Duration, error) {
	random := make([]byte, 32)
	_, err := rand.Read(random)
	if err != nil {
		return "", 0, err
	}
	refreshTokenStr := hex.EncodeToString(random)
	refreshTokenDuration := time.Duration(24*60) * time.Hour

	return refreshTokenStr, refreshTokenDuration, nil
}
