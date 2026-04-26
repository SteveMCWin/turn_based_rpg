package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"tbrpg/game"

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

	var err error
	db.Data, err = sql.Open("sqlite3", filepath.Join("data", "game.db"))
	if err != nil {
		return err
	}

	_, err = db.Data.Exec(`CREATE TABLE IF NOT EXISTS saves (
		id       INTEGER PRIMARY KEY AUTOINCREMENT,
		label    TEXT,
		saved_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		state    TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	db.is_open = true
	return nil
}

func (db *DataBase) CreateGame(g *game.Game) (int, error) {
	state, err := json.Marshal(g)
	if err != nil {
		return 0, err
	}

	result, err := db.Data.Exec(`INSERT INTO saves (state) VALUES (?)`, state)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	return int(id), err
}

func (db *DataBase) SaveGame(g *game.Game) error {
	state, err := json.Marshal(g)
	if err != nil {
		return err
	}
	_, err = db.Data.Exec(
		`UPDATE saves SET state = ?, saved_at = CURRENT_TIMESTAMP WHERE id = ?`,
		state, g.ID,
	)
	return err
}

func (db *DataBase) LoadSave(id int) (*game.Game, error) {
	var state string
	err := db.Data.QueryRow(`SELECT state FROM saves WHERE id = ?`, id).Scan(&state)
	if err != nil {
		return nil, err
	}

	var g game.Game
	if err := json.Unmarshal([]byte(state), &g); err != nil {
		return nil, err
	}

	return &g, nil
}

func (db *DataBase) ListSaves() ([]Save, error) {
	rows, err := db.Data.Query(`SELECT id, COALESCE(label, ''), saved_at FROM saves ORDER BY saved_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	saves := []Save{}
	for rows.Next() {
		var s Save
		if err := rows.Scan(&s.ID, &s.Label, &s.SavedAt); err != nil {
			return nil, err
		}
		saves = append(saves, s)
	}
	return saves, nil
}

func (db *DataBase) DeleteSave(id int) error {
	_, err := db.Data.Exec(`DELETE FROM saves WHERE id = ?`, id)
	return err
}
