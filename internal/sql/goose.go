package main

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"log"
	"io"
)

//go:embed schema/*.sql
var embedMigrations embed.FS

func Setup(db *sql.DB, path string) error {
	goose.SetVerbose(false)
	goose.SetDialect("sqlite")
	goose.SetBaseFS(embedMigrations)

	// TODO setup proper file logging here
	goose.SetLogger(log.New(io.Discard, "", 0))

	return goose.Up(db, path)
}
