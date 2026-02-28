package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pass string) (string, error) {
	hashPass, err := bcrypt.GenerateFromPassword([]byte(pass), 9)
	if err != nil {
		return "", fmt.Errorf("internal server error: failed hash password: %w", err)
	}
	return string(hashPass), nil
}

func UnHashPassword(pass string, hashPass string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(pass)); err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}