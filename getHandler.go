package erudito

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

func GetHandler(modelZero Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		modelType := reflect.ValueOf(modelZero).Type()
		modelNew := reflect.New(modelType).Interface()
		modelS := maestro.getModelStructure(modelZero)

		/*
		 * Middleware Initial
		 */
		metaData := MiddlewareMetaData{}

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_GET)
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
		 * Validation of Resource Identifier
		 */
		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		softDeleted, softDeletedOK := r.URL.Query()["del"]
		if softDeletedOK {
			if softDeleted[0] == "true" {
				db = db.Unscoped()
			}
		}

		/*
		 * Database Search for the Resource
		 */
		relString, ok := r.URL.Query()["rel"]
		if ok {
			rels := strings.Split(relString[0], ",")

			for _, rel := range rels {
				db = db.Preload(upperCamelCase(rel), func(dbPreload *gorm.DB) *gorm.DB {
					if softDeletedOK {
						if softDeleted[0] == "true" {
							dbPreload = dbPreload.Unscoped()
						}
					}

					return dbPreload
				})
			}
		}

		if notFound := db.First(modelNew, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusNotFound, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		modelGenerated, _, errR := generateReturnModel(w, r, db, modelType, modelS, reflect.ValueOf(modelNew), maestro, metaData, MIDDLEWARE_TYPE_GET, true)

		if err != nil {
			SendError(w, http.StatusForbidden, *errR)
			return
		}

		SendData(w, http.StatusOK, MakeSingularDataStruct(modelType, modelGenerated, modelS))
	})
}
