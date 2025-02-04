package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/hideckies/hermit/pkg/common/meta"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase() (*Database, error) {
	database := Database{}

	dbPath, err := meta.GetDBPath()
	if err != nil {
		return nil, err
	}

	exist := false
	if _, err := os.Stat(dbPath); err == nil {
		exist = true
	}

	database.DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if !exist {
		err := database.init()
		if err != nil {
			return &database, err
		}
	}

	return &database, nil
}

func (d *Database) init() error {
	_, err := d.DB.Exec(`
	CREATE TABLE operator (
		id INTEGER NOT NULL PRIMARY KEY,
		uuid TEXT,
		name TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = d.DB.Exec(`
	CREATE TABLE listener (
		id INTEGER NOT NULL PRIMARY KEY,
		uuid TEXT,
		name TEXT,
		protocol TEXT,
		host TEXT,
		port INTEGER,
		domains TEXT,
		active TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = d.DB.Exec(`
	CREATE TABLE agent (
		id INTEGER NOT NULL PRIMARY KEY,
		uuid TEXT,
		name TEXT,
		ip TEXT,
		os TEXT,
		arch TEXT,
		hostname TEXT,
		listener TEXT,
		sleep INTEGER,
		jitter INTEGER,
		killdate INTEGER
	)`)
	if err != nil {
		return err
	}

	return nil
}
