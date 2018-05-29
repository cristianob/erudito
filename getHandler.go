package erudito

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

func GetHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		db := DBPoolCallback(r)

		_, ok := reflect.TypeOf(model).MethodByName("BeforeGET")
		if ok {
			beforeGETr := reflect.ValueOf(model).MethodByName("BeforeGET").Call([]reflect.Value{
				reflect.ValueOf(DBPoolCallback(r)),
				reflect.ValueOf(r),
			})

			if !beforeGETr[0].Bool() {
				SendError(w, http.StatusForbidden, "Cannot access this resource!", "FORBIDDEN")
				return
			}
		}

		if db == nil {
			SendError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		relString, ok := r.URL.Query()["rel"]
		if ok {
			rels := strings.Split(relString[0], ",")

			for _, rel := range rels {
				db = db.Preload(upperCamelCase(rel))
			}
		}

		if notFound := db.First(modelNew, ModelIDField).RecordNotFound(); notFound {
			SendError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		_, ok = reflect.TypeOf(model).MethodByName("BeforeGETResponse")
		if ok {
			BeforeGETResponseR := reflect.ValueOf(model).MethodByName("BeforeGETResponse").Call([]reflect.Value{
				reflect.ValueOf(DBPoolCallback(r)),
				reflect.ValueOf(r),
				reflect.ValueOf(modelNew),
			})

			if !BeforeGETResponseR[0].Bool() {
				SendError(w, http.StatusForbidden, "Cannot access this resource!", "FORBIDDEN")
				return
			}
		}

		SendData(w, http.StatusOK, MakeSingularDataStruct(modelType, modelNew))
	})
}
