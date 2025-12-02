package pkg

import (
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(plainPassword string) (string, error)
	Verify(plainPassword, hashPassword string) bool
}

type BcryptHasher struct{}

func NewBcrypt() *BcryptHasher { return &BcryptHasher{} }

// NOTE: bcrypt only processes the first 72 bytes of the input.
// Any bytes beyond that are silently ignored.
// => Two different passwords may generate the same hash if their first 72 bytes are identical.
//
// To avoid this issue:
// - Limit password length to 72 bytes (add input validation), OR
// - Pre-hash the password using SHA-256 before passing it to bcrypt (any -> 32 bytes)
func (BcryptHasher) Hash(plainPassword string) (string, error) {
	// Pre-hash using SHA-256 to avoid 72-byte limitation in bcrypt
	sha := sha256.Sum256([]byte(plainPassword))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash), err
}

func (BcryptHasher) Verify(plainPassword, hashPassword string) bool {
	sha := sha256.Sum256([]byte(plainPassword))
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), sha[:])
	return err == nil
}
