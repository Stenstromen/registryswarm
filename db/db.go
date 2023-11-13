package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stenstromen/registryswarm/model"
	"gopkg.in/yaml.v2"
)

const createTableSQL = `CREATE TABLE IF NOT EXISTS backend_urls (
    "username" TEXT NOT NULL PRIMARY KEY,
    "scheme" TEXT,
    "host" TEXT
);`

func InitializeDatabase(dbPath string, initDataPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	// Check if the table is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM backend_urls").Scan(&count)
	if err != nil {
		return nil, err
	}

	// Load initial data if the table is empty
	if count == 0 {
		err = loadInitialData(db, initDataPath)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func loadInitialData(db *sql.DB, initDataPath string) error {
	var urls []model.BackendURL

	// Read YAML file
	file, err := os.ReadFile(initDataPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, &urls)
	if err != nil {
		return err
	}

	// Insert data into database
	for _, url := range urls {
		_, err = db.Exec("INSERT INTO backend_urls (username, scheme, host) VALUES (?, ?, ?)", url.Username, url.Scheme, url.Host)
		if err != nil {
			return err
		}
	}
	return nil
}
