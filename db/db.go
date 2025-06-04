package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// init the encrypted sqlite database
func InitDB(dbPath, dbKey string, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_key="+dbKey)
	if err != nil {
		return nil, err
	}

	// create if not exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            userid TEXT PRIMARY KEY,
            password TEXT NOT NULL
        )
    `)
	if err != nil {
		db.Close()
		return nil, err
	}

	// create test user & pass
	/*
		 *
		 * Only uncomment this if you want to test with a test account
		 * creds for test account: test:test
		 *
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
		if err != nil {
			db.Close()
			return nil, err
		}
		_, err = db.Exec("INSERT OR IGNORE INTO users (userid, password) VALUES (?, ?)", "test", string(hashedPassword))
		if err != nil {
			db.Close()
			return nil, err
		}
	*/

	logger.Info("db initialised", zap.String("path", dbPath))
	return db, nil
}

func NewUser(dbPath, dbKey string, logger *zap.Logger, userid, password string) (bool, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_key="+dbKey)
	if err != nil {
		return false, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		db.Close()
		return false, err
	}
	_, err = db.Exec("INSERT OR IGNORE INTO users (userid, password) VALUES (?, ?)", userid, string(hashedPassword))
	if err != nil {
		db.Close()
		return false, err
	}

	logger.Info("user added to database successfully", zap.String("path", dbPath))
	return true, nil
}

// integrate db
func ValidateUser(db *sql.DB, userid, password string, logger *zap.Logger) (bool, error) {
	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE userid = ?", userid).Scan(&storedPassword)
	if err == sql.ErrNoRows {
		logger.Warn("User not found", zap.String("userid", userid))
		return false, nil
	}
	if err != nil {
		logger.Error("db query failed", zap.String("userid", userid), zap.Error(err))
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		logger.Warn("invalid password", zap.String("userid", userid))
		return false, nil
	}

	return true, nil
}
