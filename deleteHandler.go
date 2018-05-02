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
			SendError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
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
			SendError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
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

		SendEmptyResponse(w, http.StatusOK)
	})
}
