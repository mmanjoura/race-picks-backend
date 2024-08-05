package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmanjoura/race-picks-backend/pkg/models"
)

type DbInstance struct {
	DB     *sql.DB
	Config map[string]string
}

var Database DbInstance

func ConnectDatabase() {

	var err error

	db, err := sql.Open("sqlite3", "./winners-ai.db")

	if err != nil {
		log.Fatal("Failed to connect to the database! \n", err)
		os.Exit(2)
	}

	configuraions, err := GetConfigs(db)
	fmt.Println(err)

	checkErr(err)

	//DB = database
	Database = DbInstance{
		DB:     db,
		Config: configuraions,
	}

}
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {

	return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)

}

func GetConfigs(db *sql.DB) (map[string]string, error) {

	ctx := context.Context(context.Background())
	rows, err := db.QueryContext(ctx, `SELECT ID,
			key,
			value

		FROM Configurations WHERE 1 = 1
		ORDER BY id ASC `,
	)
	// defer rows.Close()

	if err != nil {
		return nil, err
	}
	configurations := make(map[string]string)
	for rows.Next() {
		configuration, err := scanConfiguration(rows)

		if err != nil {
			return nil, err
		}

		if configuration.ID != 0 {
			configurations[configuration.Key] = configuration.Value
		}
	}

	return configurations, nil

}

func scanConfiguration(rows *sql.Rows) (models.Configuration, error) {
	configuration := models.Configuration{}
	err := rows.Scan(&configuration.ID,
		&configuration.Key,
		&configuration.Value,
	)

	return configuration, err
}
