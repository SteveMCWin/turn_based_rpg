package main

import (
	"net/http"
	"tbrpg/models"
	"tbrpg/handlers"
)

func main() {
	db := &models.DataBase{}
    db.InitDatabase()

	handler := handlers.SetUpRouter(db)
	http.ListenAndServe("5000", handler)
}


