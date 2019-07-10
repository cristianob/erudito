package erudito

import (
	"net/http"
	"reflect"
)

func DeleteHandler(modelZero Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "DELETE")

		modelType := reflect.ValueOf(modelZero).Type()
		modelNew := reflect.New(modelType).Interface()

		/*
		 * Middleware Initial
		 */
		metaData := MiddlewareMetaData{}

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_DELETE)
		if mwInitial.Error != nil {
			SendError(w, http.StatusForbidden, *mwInitial.Error)
			return
		}

		/*
		 * DB Connection
		 */
		db := maestro.dBPoolCallback(r, metaData)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		if notFound := db.First(modelNew, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		if err := db.Delete(modelNew).Error; err != nil {
			InternalError(w, err)
			return
		}

		SendEmptyResponse(w, http.StatusOK)
	})
}
