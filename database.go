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
	ref         string // url, link, etc
	destination string // discord channels it's sent to
	module      string
	timestamp   time.Time
}

func loadDatabase() error {
	var err error
	dbRefs, err = gorm.Open(sqlite.Open(pathDatabaseRefs), &gorm.Config{})
	if err != nil {
		return err
	}
	dbRefs.AutoMigrate(&dbRef{})

	return nil
}

func refCheckSentAnywhere(ref string) bool {
	var refs []dbRef
	dbRefs.Where("ref = ?", ref).Find(&refs)
	return len(refs) > 0
}

func refCheckSentToChannel(ref string, channel string) bool {
	var refs []dbRef
	dbRefs.Where("ref = ? AND destination = ?", ref, channel).Find(&refs)
	return len(refs) > 0
}

func refLogSent(ref string, channel string, module string) {
	dbRefs.Create(&dbRef{ref: ref, destination: channel, module: module, timestamp: time.Now()})
}