package erudito

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

type collectionCountResponse struct {
	Count uint `json:"count"`
}

func CollectionCountHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		/*
		 * Middleware Initial
		 */
		metaData := MiddlewareMetaData{}

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_POST)
		if mwInitial.Error != nil {
			SendError(w, http.StatusForbidden, *mwInitial.Error)
		}

		/*
		 * DB Connection
		 */
		db := maestro.dBPoolCallback(r, metaData)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		softDeleted, softDeletedOK := r.URL.Query()["del"]
		if softDeletedOK {
			if softDeleted[0] == "true" {
				db = db.Unscoped()
			}
		}

		if order, ok := r.URL.Query()["order"]; ok {
			db = db.Order(order[0])
		}

		if limit, ok := r.URL.Query()["limit"]; ok {
			db = db.Limit(limit[0])
		} else {
			db = db.Limit(5000)
		}

		if offset, ok := r.URL.Query()["offset"]; ok {
			db = db.Offset(offset[0])
		}

		modelNewValue := reflect.ValueOf(modelNew).Elem()
		for i := 0; i < modelNewValue.NumField(); i++ {
			fieldJSON := modelType.Field(i).Tag.Get("json")

			if getField, ok := r.URL.Query()[fieldJSON]; ok {
				for _, getFieldItem := range getField {
					getFields := strings.Split(getFieldItem, "|")

					for gi, gf := range getFields {
						switch modelType.Field(i).Type.Kind() {
						case reflect.String:
							if gi == 0 {
								db = db.Where(fieldJSON+" LIKE ?", gf)
							} else {
								db = db.Or(fieldJSON+" LIKE ?", gf)
							}

						default:
							if gi == 0 {
								db = db.Where(fieldJSON+" = ?", gf)
							} else {
								db = db.Or(fieldJSON+" = ?", gf)
							}
						}
					}
				}
			}
		}

		relString, ok := r.URL.Query()["rel"]
		if ok {
			rels := strings.Split(relString[0], ",")

			for _, rel := range rels {
				db = db.Preload(upperCamelCase(rel), func(dbPreload *gorm.DB) *gorm.DB {
					if softDeletedOK {
						if softDeleted[0] == "true" {
							dbPreload = dbPreload.Unscoped()
						}
					}

					return dbPreload
				})
			}
		}

		response := new(collectionCountResponse)
		if err := db.Model(modelNew).Count(&response.Count).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "There is an error in your query: "+err.Error(), "QUERY_ERROR")
			return
		}

		SendData(w, http.StatusOK, response)
	})
}
