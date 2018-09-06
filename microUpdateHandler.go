package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
)

type microUpdateBody struct {
	Value interface{} `json:"value"`
}

func MicroUpdateHandler(model Model, field string, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "PUT")

		beforeErrors := maestro.beforeRequestCallback(r)
		if beforeErrors != nil {
			SendError(w, 403, beforeErrors)
		}

		modelType := reflect.ValueOf(model).Type()
		var modelNew microUpdateBody
		modelDB := reflect.New(modelType).Interface()

		db := maestro.dBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&modelNew); err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type.", "")
			return
		}

		if notFound := db.First(modelDB, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		_, ok := reflect.TypeOf(model).MethodByName("ValidateField")
		if ok {
			beforePOSTr := reflect.ValueOf(model).MethodByName("ValidateField").Call([]reflect.Value{
				reflect.ValueOf(field),
				reflect.ValueOf(modelNew.Value),
				reflect.ValueOf(modelDB),
			})

			if errs := beforePOSTr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
				SendError(w, http.StatusForbidden, errs)
				return
			}
		}

		if err := db.Model(modelDB).Update(field, modelNew.Value).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(modelType, modelDB))
	})

}
