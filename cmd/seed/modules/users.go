package modules

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func SeedUsers(db *sqlx.DB, total int) []int64 {
	tx := db.MustBegin()
	defer tx.Rollback()

	pwdHash := getCommonPasswordHash("12345678")

	// 1. Prepare all user data in memory
	type userSeedData struct {
		Email    string
		Username string
	}
	pendingUsers := make([]userSeedData, total)
	for i := range pendingUsers {
		email, username := getUserSeedData(i)
		pendingUsers[i] = userSeedData{Email: email, Username: username}
	}

	// 2. Build the batch insert query
	valueStrings := make([]string, 0, total)
	valueArgs := make([]interface{}, 0, total*3)
	for i, u := range pendingUsers {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, u.Email, u.Username, pwdHash)
	}

	// 3. Execute and scan the returned IDs
	stmt := fmt.Sprintf("INSERT INTO users (email, username, password_hash) VALUES %s RETURNING id", strings.Join(valueStrings, ","))
	rows, err := tx.Query(stmt, valueArgs...)
	if err != nil {
		log.Panicf("batch insert users failed: %v", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Panicf("scan user id failed: %v", err)
		}
		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit users failed: %v", err)
	}

	log.Printf("Seeded: %d Users (e.g. email: tester@gmail.com, password: 12345678)", total)
	return ids
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
