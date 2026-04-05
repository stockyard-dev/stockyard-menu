package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-menu/internal/server"
	"github.com/stockyard-dev/stockyard-menu/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9812"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./menu-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("menu: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Menu — Self-hosted digital menu management\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("menu: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
