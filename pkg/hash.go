package pkg

import (
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword generates a bcrypt hash of the password using the default cost.
//
// To circumvent bcrypt's 72-byte input truncation limit, the password is
// pre-hashed using SHA-256 before being passed to bcrypt. This ensures
// passwords of any length are securely handled.
func HashPassword(plainText string) (string, error) {
	// SHA-256 produces a fixed 32-byte hash, safe for bcrypt.
	sha := sha256.Sum256([]byte(plainText))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash), err
}

func VerifyPassword(plainPassword, hashPassword string) bool {
	sha := sha256.Sum256([]byte(plainPassword))
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), sha[:])
	return err == nil
}
