package erudito

import (
	"net/http"
	"reflect"
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

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_COLLECTION)
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

		/*
		 * Search and Criterias
		 */
		db = getDBWithSearchCriterias(db, r, modelNew, modelType)

		response := new(collectionCountResponse)
		if err := db.Model(modelNew).Count(&response.Count).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "There is an error in your query: "+err.Error(), "QUERY_ERROR")
			return
		}

		SendData(w, http.StatusOK, response)
	})
}
