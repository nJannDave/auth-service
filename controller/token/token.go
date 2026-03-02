package token

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
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

func GetPublicKey() (*rsa.PublicKey, error) {
	pbkEnv := os.Getenv("pbk")
	pbk64, err := base64.StdEncoding.DecodeString(pbkEnv)
	if err != nil {
		return nil, errors.New("internal server error: failed decode token: " + err.Error())
	}
	pbkFix, _ := jwt.ParseRSAPublicKeyFromPEM(pbk64)
	return pbkFix, nil
}

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

func ValidateToken(token string, publicKey *rsa.PublicKey) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(tkn *jwt.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("invalid alg")
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("internal server error: failed parse token: %w", err)
	}
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := t.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("internal server error: failed type cast token: %w", err)
	}
	return claims, nil
}