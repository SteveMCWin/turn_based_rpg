package database

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DataBase struct {
	Data    *sql.DB
	is_open bool
}

type Save struct {
	ID      int    `json:"id"`
	Label   string `json:"label"`
	SavedAt string `json:"saved_at"`
}

func (db *DataBase) Close() {
	db.Data.Close()
	db.is_open = false
}

func (db *DataBase) InitDatabase() error {
	if db.is_open {
		return errors.New("database already open")
	}

	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	data_dir_name := "data"

	var err error
	db.Data, err = sql.Open("sqlite3", filepath.Join(data_dir_name, "game.db")+"?_foreign_keys=on")
	if err != nil {
		return err
	}

	sqlFiles := []string{
		filepath.Join(data_dir_name, "create_saves_table.sql"),
		filepath.Join(data_dir_name, "create_saves_table.sql"),
		filepath.Join(data_dir_name, "create_hero_tables.sql"),
		filepath.Join(data_dir_name, "create_floor_tables.sql"),
		filepath.Join(data_dir_name, "create_monster_tables.sql"),
		filepath.Join(data_dir_name, "create_shop_tables.sql"),
	}

	for _, f := range sqlFiles {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err = db.Data.Exec(string(content)); err != nil {
			return err
		}
	}

	db.is_open = true
	return nil
}
