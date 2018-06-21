package erudito

import (
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
)

func RelationRemoveHandler(model1, model2 Model, fieldName string, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "DELETE")

		model1Type := reflect.ValueOf(model1).Type()
		model2Type := reflect.ValueOf(model2).Type()

		model1DB := reflect.New(model1Type).Interface()
		model2DB := reflect.New(model2Type).Interface()

		db := DBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		Model1IDField, err := GetNumericRouteField(r, "id1")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "First entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		Model2IDField, err := GetNumericRouteField(r, "id2")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Second entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		if notFound := db.First(model1DB, Model1IDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "First entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		if notFound := db.First(model2DB, Model2IDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Second entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		if err := db.Model(model1DB).Association(fieldName).Delete(model2DB).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(model1Type, model1DB))
	})

}
