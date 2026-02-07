package modules

import (
	"crypto/sha256"
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func SeedUsers(db *sqlx.DB, total int) []int64 {
	hashedPassword, err := hashPassword("12345678")
	if err != nil {
		log.Panic("hash password error")
	}

	ids := make([]int64, 0, total)
	tx := db.MustBegin()

	for range total {
		var id int64
		err := tx.Get(&id, `
            INSERT INTO users (email, username, password_hash)
            VALUES ($1, $2, $3)
            RETURNING id
        `,
			gofakeit.Email(),
			gofakeit.Username(),
			hashedPassword,
		)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
		ids = append(ids, id)
	}

	tx.Commit()

	return ids
}

func TruncateUser(db *sqlx.DB) {
	_, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Panicf("cannot clean data: %v", err)
	}
}

func hashPassword(plainText string) (string, error) {
	sha := sha256.Sum256([]byte(plainText))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash), err
}
