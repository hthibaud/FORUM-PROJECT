package data

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./database.db"

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base de données : %w", err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        email TEXT NOT NULL
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("erreur lors de la création des tables : %w", err)
	}

	fmt.Printf("Base de données '%s' bien créée.\n", dbPath)
	return db, nil
}
