package cmd

import (
	"database/sql"

	"github.com/spf13/viper"
)

func openDB() *sql.DB {
	db, err := sql.Open("sqlite", viper.GetString(DB_PATH))
	check(err)
	return db
}
