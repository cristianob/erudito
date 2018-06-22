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

func CollectionCountHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		db := DBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
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
				switch modelType.Field(i).Type.Kind() {
				case reflect.String:
					db = db.Where(fieldJSON+" LIKE ?", getField)

				case reflect.Int, reflect.Uint:
					db = db.Where(fieldJSON+" = ?", getField)
				}
			}
		}

		relString, ok := r.URL.Query()["rel"]
		if ok {
			rels := strings.Split(relString[0], ",")

			for _, rel := range rels {
				db = db.Preload(upperCamelCase(rel))
			}
		}

		softDeleted, ok := r.URL.Query()["del"]
		if ok {
			if softDeleted[0] == "true" {
				db = db.Unscoped()
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
