package main

import (
	"log"
	"net/http"
	"tbrpg/game"
	"tbrpg/handlers"
	"tbrpg/database"
)

func main() {
	config, err := game.LoadConfig("config")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db := &database.DataBase{}
	if err := db.InitDatabase(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	log.Println("Starting server on :5000")
	if err := http.ListenAndServe(":5000", handlers.NewServer(config, db)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}


