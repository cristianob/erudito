package main

import (
	"log"
	"net/http"

	"github.com/cristianob/erudito"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var dbConn *gorm.DB

func databaseInit() {
	var err error

	dbConn, err = gorm.Open("mysql", "root:erudito@tcp(db:3306)/erudito?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("[TEST] Database connection failed: " + err.Error())
	}
}

func databaseMigrate() {
	dbConn.AutoMigrate(
		&A{},
		&B{},
		&C{},
		&D{},
		&E{},
	)
}

func databaseResolve(r *http.Request, metaData erudito.MiddlewareMetaData) *gorm.DB {
	log.Println("Resolve received: ", metaData)
	return dbConn
}

func databaseClose() {
	dbConn.Close()
}
