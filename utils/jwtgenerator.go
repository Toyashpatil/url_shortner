package utils

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJwtSecret() []byte {
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	return []byte(os.Getenv("JWT_SECRET"))
}

func GetJwtToken(userId string) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"client":     "eventio",
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
		"iat":        1610000000,
		"iss":        "eventio",
	}
	claims["user_id"] = userId
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJwtSecret())
}
