package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"

	"bitbucket.org/laticin/coleta/helpers"
	"github.com/jinzhu/gorm"
)

func PutHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelType := reflect.ValueOf(model).Type()
		modelSend := reflect.New(modelType).Interface()
		modelDB := reflect.New(modelType).Interface()

		db := DBPoolCallback(r)
		if db == nil {
			SendError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		if err := json.NewDecoder(r.Body).Decode(modelSend); err != nil {
			SendError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type.", "")
			return
		}

		if notFound := db.First(modelDB, ModelIDField).RecordNotFound(); notFound {
			SendError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		modelDBValue := reflect.ValueOf(modelDB).Elem()
		modelSendValue := reflect.ValueOf(modelSend).Elem()
		for i := 0; i < modelDBValue.NumField(); i++ {
			fieldName := modelType.Field(i).Name

			switch modelType.Field(i).Type.Kind() {
			case reflect.Slice:
				modelDBValue.Field(i).Set(modelSendValue.FieldByName(fieldName))
				break
			default:
				if modelType.Field(i).Name != "FullModel" && modelType.Field(i).Name != "NoDeleteModel" {
					if modelSendValue.FieldByName(fieldName).Interface() != reflect.Zero(modelType.Field(i).Type).Interface() {
						modelDBValue.Field(i).Set(modelSendValue.FieldByName(fieldName))
					}
				}
				break
			}
		}

		if err := db.Save(modelDB).Error; err != nil {
			helpers.SendError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		SendData(w, MakeSingularDataStruct(modelType, modelDB))
	})
}
