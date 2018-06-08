package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
)

func PutHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()
		modelDB := reflect.New(modelType).Interface()

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

		if err := json.NewDecoder(r.Body).Decode(modelNew); err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type.", "")
			return
		}

		validationErrors := []JSendErrorDescription{}
		for i := 0; i < modelType.NumField(); i++ {
			if len(modelType.Field(i).PkgPath) != 0 {
				continue
			}

			if !checkIfTagExists("excludePUT", modelType.Field(i).Tag.Get("erudito")) {
				beforePOSTr := reflect.ValueOf(model).MethodByName("ValidateField").Call([]reflect.Value{
					reflect.ValueOf(modelType.Field(i).Tag.Get("json")),
					reflect.ValueOf(modelNew).Elem().Field(i),
					reflect.ValueOf(modelNew),
				})

				if errs := beforePOSTr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
					validationErrors = append(validationErrors, errs...)
				}
			}
		}

		if len(validationErrors) > 0 {
			SendError(w, http.StatusForbidden, validationErrors)
			return
		}

		for i := 0; i < modelType.NumField(); i++ {
			for _, prop := range getTagValues("PUT", modelType.Field(i).Tag.Get("erudito")) {
				switch prop {
				case "auto_remove":
					db = db.Preload(modelType.Field(i).Name)
				}
			}
		}

		if notFound := db.First(modelDB, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		modelDBValue := reflect.ValueOf(modelDB)
		modelNewValue := reflect.ValueOf(modelNew)

		for i := 0; i < modelType.NumField(); i++ {
			for _, prop := range getTagValues("PUT", modelType.Field(i).Tag.Get("erudito")) {
				switch prop {
				case "auto_remove":
					setDB := modelDBValue.Elem().FieldByName(modelType.Field(i).Name)
					setNew := modelNewValue.Elem().FieldByName(modelType.Field(i).Name)
					setDiference := []reflect.Value{}

					for j := 0; j < setDB.Len(); j++ {
						exists := false
						for k := 0; k < setNew.Len(); k++ {
							if setDB.Index(j).FieldByName("ID").Interface().(uint) == setNew.Index(k).FieldByName("ID").Interface().(uint) {
								exists = true
							}
						}

						if !exists {
							setDiference = append(setDiference, setDB.Index(j))
						}
					}

					for _, modelRemove := range setDiference {
						db2 := DBPoolCallback(r)
						db2.Delete(modelRemove.Interface())
					}
				}
			}
		}

		deepCopy(modelType, modelNewValue.Elem(), modelDBValue.Elem(), "excludePUT")

		_, ok := reflect.TypeOf(model).MethodByName("BeforePUT")
		if ok {
			beforePUT := reflect.ValueOf(model).MethodByName("BeforePUT").Call([]reflect.Value{
				reflect.ValueOf(DBPoolCallback(r)),
				reflect.ValueOf(r),
				modelDBValue,
			})

			if errs := beforePUT[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
				SendError(w, http.StatusForbidden, errs)
				return
			}
		}

		if err := db.Save(modelDB).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		/*
		 * Removal of the ExcludeGET fields
		 */
		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("excludeGET", modelType.Field(i).Tag.Get("erudito")) {
				reflect.ValueOf(modelDB).Elem().Field(i).Set(reflect.Zero(modelType.Field(i).Type))
			}
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(modelType, modelDB))
	})

}
