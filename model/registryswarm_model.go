package model

import (
	"database/sql"
	"log"
	"net/url"
)

type BackendURL struct {
	Username string `yaml:"username"`
	Scheme   string `yaml:"scheme"`
	Host     string `yaml:"host"`
}

type Registry struct {
	Username string `json:"username"`
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
}

func GetBackendURL(db *sql.DB, username string) (*url.URL, error) {
	var scheme, host string
	err := db.QueryRow("SELECT scheme, host FROM backend_urls WHERE username = ?", username).Scan(&scheme, &host)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &url.URL{Scheme: scheme, Host: host}, nil
}

func GetRegistries(db *sql.DB) ([]Registry, error) {
	const query = "SELECT username, scheme, host FROM backend_urls"

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying registries: %v", err)
		return nil, err
	}
	defer rows.Close()

	var registries []Registry
	for rows.Next() {
		var reg Registry
		if err := rows.Scan(&reg.Username, &reg.Scheme, &reg.Host); err != nil {
			log.Printf("Error scanning registry: %v", err)
			return nil, err
		}
		registries = append(registries, reg)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil, err
	}

	return registries, nil
}
