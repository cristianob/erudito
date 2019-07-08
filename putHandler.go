package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func PutHandler(modelZero Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "PUT")

		/*
		 * Getting ID field
		 */
		ModelIDField, errN := GetNumericRouteField(r, "id")
		if errN != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

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

		/*
		 * JSON Unmarshal
		 */
		var modelUnmarshal map[string]interface{}

		if err := json.NewDecoder(r.Body).Decode(&modelUnmarshal); err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Given request body is invalid: "+err.Error(), "")
			return
		}

		modelUnmarshal["id"] = ModelIDField

		/*
		 * Generate model and run PRE Middlewares
		 */
		modelType := reflect.ValueOf(modelZero).Type()
		modelS := maestro.getModelStructure(modelZero)
		modelGenerated, metaData, err := generatePostModel(w, r, db, modelType, modelS, modelUnmarshal, maestro, metaData, MIDDLEWARE_TYPE_PUT, false)

		if err != nil {
			SendError(w, http.StatusForbidden, *err)
			return
		}

		/*
		 * Insert in DB
		 */
		if err := db.Save(modelGenerated).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot create a new record - "+err.Error(), "ENTITY_CREATE_ERROR")
			return
		}

		/*
		 * Inserting Multiple relation IDs
		 */
		modelGenerated, err = insertMultipleRelations(w, r, db, modelType, modelS, modelUnmarshal, reflect.ValueOf(modelGenerated).Elem(), maestro, true)

		if err != nil {
			SendError(w, http.StatusForbidden, *err)
			return
		}

		/*
		 * Generate return and run POS Middlewares
		 */
		modelGenerated, _, err = generateReturnModel(w, r, db, modelType, modelS, reflect.ValueOf(modelGenerated), maestro, metaData, MIDDLEWARE_TYPE_PUT, true)

		if err != nil {
			SendError(w, http.StatusForbidden, *err)
			return
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(modelType, modelGenerated, modelS))
	})
}
