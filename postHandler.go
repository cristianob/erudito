package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
)

func PostHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := DBPoolCallback(r)
		if db == nil {
			SendError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		if err := json.NewDecoder(r.Body).Decode(modelNew); err != nil {
			SendError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type.", "")
			return
		}

		db.NewRecord(modelNew)
		if err := db.Create(modelNew).Error; err != nil {
			SendError(w, http.StatusForbidden, "Cannot create a new record - "+err.Error(), "ENTITY_CREATE_ERROR")
			return
		}

		SendData(w, MakeSingularDataStruct(modelType, modelNew))
	})
}
