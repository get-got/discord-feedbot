package main

import (
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	pathDatabaseRefs = pathData + string(os.PathSeparator) + "reference-log.db"
)

var (
	dbRefs *gorm.DB
)

type dbRef struct {
	gorm.Model
	ref       string // url, link, etc
	module    string
	sentTo    []string // discord channels it's sent to
	timestamp time.Time
}

func loadDatabase() error {
	dbRefs, err := gorm.Open(sqlite.Open(pathDatabaseRefs), &gorm.Config{})
	if err != nil {
		return err
	}
	dbRefs.AutoMigrate(&dbRef{})

	return nil
}
