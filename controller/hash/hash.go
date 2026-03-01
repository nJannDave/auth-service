package hash

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pass string) (string, error) {
	hashPass, err := bcrypt.GenerateFromPassword([]byte(pass), 9)
	if err != nil {
		return "", errors.New("internal server error: failed hash password: " + err.Error())
	}
	return string(hashPass), nil
}

func UnHashPassword(pass string, hashPass string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(pass)); err != nil {
		return errors.New("invalid password")
	}
	return nil
}