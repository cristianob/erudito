package main

import (
	"net/http"

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

func databaseResolve(r *http.Request) *gorm.DB {
	return dbConn
}

func databaseClose() {
	dbConn.Close()
}
