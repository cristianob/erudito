package erudito

import (
	"net/http"
	"reflect"
)

func CollectionHandler(modelZero Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		modelType := reflect.ValueOf(modelZero).Type()
		modelNew := reflect.New(modelType).Interface()
		modelS := maestro.getModelStructure(modelZero)

		/*
		 * Middleware Initial
		 */
		metaData := MiddlewareMetaData{}

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_COLLECTION)
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

		/*
		 * Search and Criterias
		 */
		db = getDBWithSearchCriterias(db, r, modelNew, modelType)

		modelSlice := reflect.New(reflect.SliceOf(modelType))
		if err := db.Find(modelSlice.Interface()).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "There is an error in your query: "+err.Error(), "QUERY_ERROR")
			return
		}

		modelSlice = modelSlice.Elem()

		/*
		 *  MiddlewareAfter
		 */
		rtrSlice := []interface{}{}
		for i := 0; i < modelSlice.Len(); i++ {
			modelGenerated, _, err := generateReturnModel(w, r, db, modelType, modelS, modelSlice.Index(i), maestro, metaData, MIDDLEWARE_TYPE_COLLECTION, true)

			if err != nil {
				SendError(w, http.StatusForbidden, *err)
				return
			}

			rtrSlice = append(rtrSlice, modelGenerated)
		}

		SendData(w, http.StatusOK, MakeArrayDataStruct(modelType, rtrSlice, modelS))
	})
}
