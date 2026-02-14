package modules

import (
	"crypto/sha256"
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func SeedUsers(db *sqlx.DB, total int) []int64 {
	tx := db.MustBegin()
	defer tx.Rollback()

	stmt := prepareUserStmt(tx)
	pwdHash := getCommonPasswordHash("12345678")
	ids := make([]int64, total)

	for i := range total {
		email, username := getUserSeedData(i)

		var id int64
		if err := stmt.QueryRow(email, username, pwdHash).Scan(&id); err != nil {
			log.Panicf("insert user [%s] failed: %v", email, err)
		}
		ids[i] = id
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit users failed: %v", err)
	}

	log.Printf("Seeded: %d Users (e.g. email: tester@gmail.com, password: 12345678)", total)
	return ids
}

// --- Internal Helpers ---

func prepareUserStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO users (email, username, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id
    `)
	if err != nil {
		log.Panicf("prepare user stmt failed: %v", err)
	}
	return stmt
}

func getUserSeedData(index int) (string, string) {
	if index == 0 {
		return "tester@gmail.com", "tester"
	}
	return gofakeit.Email(), gofakeit.Username()
}

func getCommonPasswordHash(password string) string {
	hash, err := hashPassword(password)
	if err != nil {
		log.Panicf("hash password failed: %v", err)
	}
	return hash
}

// --- Utilities ---

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
