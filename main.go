package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stenstromen/registryswarm/controller"
	"github.com/stenstromen/registryswarm/db"
)

const APIVersion = "/v1"

func main() {
	dbPath := "database.db"
	initDataPath := "registries.yaml"

	db, err := db.InitializeDatabase(dbPath, initDataPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", controller.ProxyRequest(db))
	http.HandleFunc(APIVersion+"/controller", controller.GetRegistries(db))

	fmt.Println("Starting proxy server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting proxy server:", err)
	}
}
