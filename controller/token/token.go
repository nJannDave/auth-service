package token

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Token struct {
	Access string
	Refresh string
}

var prk *rsa.PrivateKey

func Init() error {
	prkEnv := os.Getenv("prk")
	prk64, err := base64.StdEncoding.DecodeString(prkEnv)
	if err != nil {
		return errors.New("internal server error: failed decode token: " + err.Error())
	}
	prkFix, _ := jwt.ParseRSAPrivateKeyFromPEM(prk64)
	prk = prkFix
	return nil
}

func generateToken(id int, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		ID: id,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
            IssuedAt: jwt.NewNumericDate(time.Now()),
        },
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tknStr, err := token.SignedString(prk)
	if err != nil {
		return "", errors.New("internal server error: failed signed token: " + err.Error())
	}
	return tknStr, nil
}

func GenerateToken(id int, role string) (*Token, error) {
	a, err := generateToken(id, role, 3*time.Minute)
	if err != nil {
		return nil, err
	}
	r, err := generateToken(id, role, 24*3*time.Hour)
	if err != nil {
		return nil, err
	}
	return &Token{
		Access: a,
		Refresh: r,
	}, nil
}