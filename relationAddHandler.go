package erudito

import (
	"net/http"
	"reflect"
)

func RelationAddHandler(model1, model2 Model, fieldName string, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "PUT")

		model1Type := reflect.ValueOf(model1).Type()
		model2Type := reflect.ValueOf(model2).Type()
		modelS := maestro.getModelStructure(model1)

		model1DB := reflect.New(model1Type).Interface()
		model2DB := reflect.New(model2Type).Interface()

		/*
		 * Middleware Initial
		 */
		metaData := MiddlewareMetaData{}

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_RELATION)
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

		if err := db.Model(model1DB).Association(fieldName).Append(model2DB).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		modelGenerated, _, errR := generateReturnModel(w, r, db, model1Type, modelS, reflect.ValueOf(model1DB), maestro, metaData, MIDDLEWARE_TYPE_RELATION, true)

		if err != nil {
			SendError(w, http.StatusForbidden, *errR)
			return
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(model1Type, modelGenerated, modelS))
	})

}
