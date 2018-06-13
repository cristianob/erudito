package erudito

import (
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
)

func DeleteHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		db := DBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		for i := 0; i < modelType.NumField(); i++ {
			for _, prop := range getTagValues("DELETE", modelType.Field(i).Tag.Get("erudito")) {
				switch prop {
				case "auto_remove":
					db = db.Preload(modelType.Field(i).Name)
				}
			}
		}

		if notFound := db.First(modelNew, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		_, ok := reflect.TypeOf(model).MethodByName("BeforeDELETE")
		if ok {
			beforeDELETE := reflect.ValueOf(model).MethodByName("BeforeDELETE").Call([]reflect.Value{
				reflect.ValueOf(DBPoolCallback(r)),
				reflect.ValueOf(r),
				reflect.ValueOf(modelNew),
			})

			if errs := beforeDELETE[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
				SendError(w, http.StatusForbidden, errs)
				return
			}
		}

		modelNewValue := reflect.ValueOf(modelNew)

		for i := 0; i < modelType.NumField(); i++ {
			for _, prop := range getTagValues("DELETE", modelType.Field(i).Tag.Get("erudito")) {
				switch prop {
				case "auto_remove":
					setNew := modelNewValue.Elem().FieldByName(modelType.Field(i).Name)
					setDelete := []reflect.Value{}

					for j := 0; j < setNew.Len(); j++ {
						setDelete = append(setDelete, setNew.Index(j))
					}

					for _, modelRemove := range setDelete {
						db2 := DBPoolCallback(r)
						db2.Delete(modelRemove.Interface())
					}
				}
			}
		}

		if err := db.Delete(modelNew).Error; err != nil {
			InternalError(w, err)
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization")
		w.Header().Add("Access-Control-Max-Age", "1728000")

		SendEmptyResponse(w, http.StatusOK)
	})
}
