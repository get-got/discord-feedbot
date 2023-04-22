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
	Ref       string // url, link, etc
	Channel   string // discord channel it's sent to
	Module    string
	Timestamp time.Time
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
	dbRefs.Model(&dbRef{}).Where("`ref` = ?", ref).Find(&refs)
	return len(refs) > 0
}

func refCheckSentToChannel(ref string, channel string) bool {
	var refs []dbRef
	dbRefs.Model(&dbRef{}).Where("`channel` = ? AND `ref` = ?", channel, ref).Find(&refs)
	return len(refs) > 0
}

func refLogSent(ref string, channel string, module string) {
	dbRefs.Create(&dbRef{Ref: ref, Channel: channel, Module: module, Timestamp: time.Now()})
}
