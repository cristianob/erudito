package erudito

import (
	"net/http"
	"reflect"
)

func DeleteHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "DELETE")

		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

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

		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("DELETEautoremove", modelType.Field(i).Tag.Get("erudito")) || checkIfTagExists("DELETEautodelete", modelType.Field(i).Tag.Get("erudito")) {
				db = db.Preload(modelType.Field(i).Name)
			}
		}

		if notFound := db.First(modelNew, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		_, ok := reflect.TypeOf(model).MethodByName("BeforeDELETE")
		if ok {
			beforeDELETE := reflect.ValueOf(model).MethodByName("BeforeDELETE").Call([]reflect.Value{
				reflect.ValueOf(maestro.dBPoolCallback(r, metaData)),
				reflect.ValueOf(r),
				reflect.ValueOf(modelNew),
				reflect.ValueOf(metaData),
			})

			if errs := beforeDELETE[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
				SendError(w, http.StatusForbidden, errs)
				return
			}
		}

		modelNewValue := reflect.ValueOf(modelNew)

		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("DELETEautoremove", modelType.Field(i).Tag.Get("erudito")) {
				setNew := modelNewValue.Elem().FieldByName(modelType.Field(i).Name)
				setDelete := []reflect.Value{}

				for j := 0; j < setNew.Len(); j++ {
					setDelete = append(setDelete, setNew.Index(j))
				}

				for _, modelRemove := range setDelete {
					db2 := maestro.dBPoolCallback(r, metaData)
					db2.Model(modelNew).Association(modelType.Field(i).Name).Delete(modelRemove.Interface())
				}
			}

			if checkIfTagExists("DELETEautodelete", modelType.Field(i).Tag.Get("erudito")) {
				setNew := modelNewValue.Elem().FieldByName(modelType.Field(i).Name)
				setDelete := []reflect.Value{}

				for j := 0; j < setNew.Len(); j++ {
					setDelete = append(setDelete, setNew.Index(j))
				}

				for _, modelRemove := range setDelete {
					db2 := maestro.dBPoolCallback(r, metaData)
					db2.Delete(modelRemove.Interface())
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
