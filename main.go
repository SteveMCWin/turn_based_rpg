package main

import (
	"log"
	"net/http"
	"tbrpg/game"
	"tbrpg/handlers"
	"tbrpg/models"
)

func main() {
	config, err := game.LoadConfig("config")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db := &models.DataBase{}
    db.InitDatabase()

	log.Println("Starting server on :5000")
	if err := http.ListenAndServe(":5000", handlers.NewServer(config, db.Data)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}


