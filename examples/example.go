package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/h44z/lightmigrate"
	"github.com/h44z/lightmigrate-mysql/mysql"
)

func main() {
	sqlClient, err := getSqlClient("username:password@tcp(123.123.123.123)/migration_test_db?multiStatements=true")
	if err != nil {
		log.Fatalf("unable to setup sql client: %v", err)
	}
	defer sqlClient.Close()

	fsys := os.DirFS("examples")
	source, err := lightmigrate.NewFsSource(fsys, "migrations")
	if err != nil {
		log.Fatalf("unable to setup source: %v", err)
	}
	defer source.Close()

	driver, err := mysql.NewDriver(sqlClient, "migration_test_db")
	if err != nil {
		log.Fatalf("unable to setup driver: %v", err)
	}
	defer driver.Close()

	migrator, err := lightmigrate.NewMigrator(source, driver, lightmigrate.WithVerboseLogging(true))
	if err != nil {
		log.Fatalf("unable to setup migrator: %v", err)
	}

	err = migrator.Migrate(1) // Migrate to schema version 1
	if err != nil {
		log.Fatalf("migration error: %v", err)
	}
}

func getSqlClient(url string) (*sql.DB, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	return db, nil
}
