package db

import (
	"Forum/pkg/utils"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

const dbPath = "./data/database.db"

var db *sql.DB

func Init() {
	idDbNouvelle := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		idDbNouvelle = true
		utils.Log("The database file does not exist. It will be created.")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if idDbNouvelle {
		utils.Debug("Initializing a Blank Database...")
	} else {
		utils.Debug("The file already exists. Checking the integrity of the tables...")
	}

	for nomTable, requeteCreation := range schema {
		existe, err := tableExiste(db, nomTable)
		if err != nil {
			log.Fatalf("Error when checking the table %s: %v", nomTable, err)
		}

		if !existe {
			utils.Debug(fmt.Sprintf("The table '%s' is missing. Creating...\n", nomTable))
			_, err := db.Exec(requeteCreation)
			if err != nil {
				log.Fatalf("Error creating the table %s: %v", nomTable, err)
			}
			utils.Debug(fmt.Sprintf("Table '%s' created successfully.\n", nomTable))
		} else {
			utils.Debug(fmt.Sprintf("Table '%s' already present. RAS.\n", nomTable))
		}
	}

	utils.Log("[OK] Database ready and verified.")
}

func tableExiste(db *sql.DB, tableName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`

	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil // La table n'existe pas
	}
	if err != nil {
		return false, err // Une vraie erreur s'est produite
	}

	return true, nil // La table existe
}
