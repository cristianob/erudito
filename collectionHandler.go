package erudito

import (
	"net/http"
	"reflect"
	"strings"

	"bitbucket.org/laticin/coleta/helpers"
	"github.com/jinzhu/gorm"
)

func CollectionHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		db := DBPoolCallback(r)
		if db == nil {
			SendError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
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
					break

				case reflect.Uint:
				case reflect.Int:
					db = db.Where(fieldJSON+" = ?", getField)
					break

				default:
					break
				}
			}
		}

		relString, ok := r.URL.Query()["rel"]
		if ok {
			rels := strings.Split(relString[0], ",")

			for _, rel := range rels {
				db = db.Preload(strings.Title(rel))
			}
		}

		modelSlice := reflect.New(reflect.SliceOf(modelType)).Interface()
		if err := db.Find(modelSlice).Error; err != nil {
			helpers.SendError(w, http.StatusForbidden, "There is an error in your query: "+err.Error(), "QUERY_ERROR")
			return
		}

		SendData(w, MakeArrayDataStruct(modelType, modelSlice))
	})
}
