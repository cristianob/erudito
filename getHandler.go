package erudito

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

func GetHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		var metaData map[string]interface{}
		var beforeErrors []JSendErrorDescription
		if maestro.beforeRequestCallback != nil {
			beforeErrors, metaData = maestro.beforeRequestCallback(r)
			if beforeErrors != nil {
				SendError(w, 403, beforeErrors)
				return
			}
		}

		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		// Database connection
		db := maestro.dBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		/*
		 * BeforeGET Handler
		 */
		_, ok := reflect.TypeOf(model).MethodByName("BeforeGET")
		if ok {
			beforeGETr := reflect.ValueOf(model).MethodByName("BeforeGET").Call([]reflect.Value{
				reflect.ValueOf(db),
				reflect.ValueOf(r),
				reflect.ValueOf(metaData),
			})

			if errs := beforeGETr[1].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
				SendError(w, http.StatusForbidden, errs)
				return
			}

			db = beforeGETr[0].Interface().(*gorm.DB)
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

		/*
		 * BeforeGETResponse Handler
		 */
		_, ok = reflect.TypeOf(model).MethodByName("BeforeGETResponse")
		if ok {
			beforeGETResponseR := reflect.ValueOf(model).MethodByName("BeforeGETResponse").Call([]reflect.Value{
				reflect.ValueOf(maestro.dBPoolCallback(r)),
				reflect.ValueOf(r),
				reflect.ValueOf(modelNew),
				reflect.ValueOf(metaData),
			})

			if errs := beforeGETResponseR[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
				SendError(w, http.StatusForbidden, errs)
				return
			}
		}

		/*
		 * Removal of the ExcludeGET fields
		 */
		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("excludeGET", modelType.Field(i).Tag.Get("erudito")) {
				reflect.ValueOf(modelNew).Elem().Field(i).Set(reflect.Zero(modelType.Field(i).Type))
			}
		}

		SendData(w, http.StatusOK, MakeSingularDataStruct(modelType, modelNew))
	})
}
