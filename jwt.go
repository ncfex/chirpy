package main

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserJWTPayload struct {
	Id int
}

const DEFAULT_EXPIRES_DURATION = time.Hour * 24

func (cfg *apiConfig) GenerateJWT(issuer string, payload UserJWTPayload, expiresInSeconds time.Duration) (string, error) {
	issuedAt := time.Now().UTC()
	durationToUse := DEFAULT_EXPIRES_DURATION

	if expiresInSeconds != 0 && expiresInSeconds <= DEFAULT_EXPIRES_DURATION {
		durationToUse = expiresInSeconds
	}
	expiresAt := issuedAt.Add(durationToUse)

	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Subject:   strconv.Itoa(payload.Id),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(cfg.jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
