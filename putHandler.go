package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func PutHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "PUT")

		if maestro.beforeRequestCallback != nil {
			beforeErrors := maestro.beforeRequestCallback(r)
			if beforeErrors != nil {
				SendError(w, 403, beforeErrors)
				return
			}
		}

		modelType := reflect.ValueOf(model).Type()

		modelNewValue := reflect.New(modelType)
		modelNew := modelNewValue.Interface()

		modelDBValue := reflect.New(modelType)
		modelDB := modelDBValue.Interface()

		/*
		 * DB Connection
		 */
		db := maestro.dBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		/*
		 * Getting ID field
		 */
		ModelIDField, err := GetNumericRouteField(r, "id")
		if err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Entity ID is invalid", "ENTITY_ID_INVALID")
			return
		}

		/*
		 * JSON Unmarshal
		 */
		if err := json.NewDecoder(r.Body).Decode(modelNew); err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type.", "")
			return
		}

		/*
		 * Treating autoRemove fields
		 */

		// Preloading fields
		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("PUTautoremove", modelType.Field(i).Tag.Get("erudito")) {
				db = db.Preload(modelType.Field(i).Name)
			}
		}

		// Getting DB model
		if notFound := db.First(modelDB, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		// Getting diff and removing
		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("PUTautoremove", modelType.Field(i).Tag.Get("erudito")) {
				setDB := modelDBValue.Elem().FieldByName(modelType.Field(i).Name)
				setNew := modelNewValue.Elem().FieldByName(modelType.Field(i).Name)
				setDiference := []reflect.Value{}

				for j := 0; j < setDB.Len(); j++ {
					exists := false
					for k := 0; k < setNew.Len(); k++ {
						dbIndex := setDB.Index(j)
						if dbIndex.Kind() == reflect.Ptr {
							dbIndex = dbIndex.Elem()
						}

						newIndex := setNew.Index(k)
						if newIndex.Kind() == reflect.Ptr {
							newIndex = newIndex.Elem()
						}

						if dbIndex.FieldByName("ID").Interface().(uint) == newIndex.FieldByName("ID").Interface().(uint) {
							exists = true
						}
					}

					if !exists {
						setDiference = append(setDiference, setDB.Index(j))
					}
				}

				for _, modelRemove := range setDiference {
					db2 := maestro.dBPoolCallback(r)
					db2.Delete(modelRemove.Interface())
				}
			}
		}

		/*
		 * Recursive Validation and Cleaning
		 */
		validationErrors := validateAndClearPUT(modelType, modelNewValue, maestro, r, []string{}, nil)

		if len(validationErrors) > 0 {
			SendError(w, http.StatusForbidden, validationErrors)
			return
		}

		/*
		 * Get the ID and Erudito's reserved fields
		 */
		modelNewValue.Elem().FieldByName("ID").Set(modelDBValue.Elem().FieldByName("ID"))

		if _, ok := modelType.FieldByName("CreatedAt"); ok {
			modelNewValue.Elem().FieldByName("CreatedAt").Set(modelDBValue.Elem().FieldByName("CreatedAt"))
		}

		if _, ok := modelType.FieldByName("DeletedAt"); ok {
			modelNewValue.Elem().FieldByName("DeletedAt").Set(modelDBValue.Elem().FieldByName("DeletedAt"))
		}

		/*
		 * Saving the new model
		 */
		if err := db.Save(modelNew).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		/*
		 * Removal of the ExcludeGET fields
		 */
		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("excludeGET", modelType.Field(i).Tag.Get("erudito")) {
				reflect.ValueOf(modelNew).Elem().Field(i).Set(reflect.Zero(modelType.Field(i).Type))
			}
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(modelType, modelNew))
	})

}

func validateAndClearPUT(model reflect.Type, source reflect.Value, maestro *maestro, r *http.Request, stack []string, slicePos *uint) []JSendErrorDescription {
	// Final array of errors
	validationErrors := []JSendErrorDescription{}

	// Call BeforePUT in recursion
	_, hasBeforePUT := model.MethodByName("BeforePUT")
	if hasBeforePUT {
		beforePUTr := source.MethodByName("BeforePUT").Call([]reflect.Value{
			reflect.ValueOf(maestro.dBPoolCallback(r)),
			reflect.ValueOf(r),
			source,
		})

		if errs := beforePUTr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
			validationErrors = append(validationErrors, errs...)
		}
	}

	// Iterate through Model Fields
	for i := 0; i < model.NumField(); i++ {
		// If have a Package Path, go to next
		if len(model.Field(i).PkgPath) != 0 {
			continue
		}

		// If it's a Erudito's reserved field
		if model.Field(i).Name == "ID" || model.Field(i).Name == "CreatedAt" || model.Field(i).Name == "UpdatedAt" || model.Field(i).Name == "DeletedAt" {
			continue
		}

		// If is a pointer, we do the dereference
		if source.Type().Kind() == reflect.Ptr {
			source = source.Elem()
		}

		// We store a boolean to see if this model has a Validate method
		_, hasValidateField := model.MethodByName("ValidateField")

		// Switching through the field's Kind
		switch model.Field(i).Type.Kind() {

		// If is a struct
		case reflect.Struct:
			// We verify if is a Erudito model or is other field like time.Time
			if model.Field(i).Type.Implements(reflect.TypeOf((*Model)(nil)).Elem()) ||
				model.Field(i).Type == reflect.TypeOf(FullModel{}) ||
				model.Field(i).Type == reflect.TypeOf(HardDeleteModel{}) ||
				model.Field(i).Type == reflect.TypeOf(SimpleModel{}) {
				// If is a Erudito model, whe do recursion
				validationErrors = append(validationErrors, validateAndClearPUT(model.Field(i).Type, source.Field(i).Addr(), maestro, r, append(stack, model.Field(i).Tag.Get("json")), nil)...)
			} else {
				// If not, whe validate the field (if the model has the function)
				if hasValidateField {
					validationErrors = append(validationErrors, validateField(model.Field(i), source.Field(i), source.Addr(), stack, slicePos)...)
				}
			}

		// If is a slice
		case reflect.Slice:
			// We iterate throught the slice doing recursion
			for j := 0; j < source.Field(i).Len(); j++ {
				pos := uint(j)

				if source.Field(i).Index(j).Kind() == reflect.Struct {
					// If is a struct, we do recursion
					validationErrors = append(validationErrors, validateAndClearPUT(source.Field(i).Index(j).Type(), source.Field(i).Index(j).Addr(), maestro, r, append(stack, model.Field(i).Tag.Get("json")), &pos)...)
				} else {
					// If not, whe validate the field (if the model has the function)
					if hasValidateField {
						validationErrors = append(validationErrors, validateField(model.Field(i), source.Field(i), source.Addr(), stack, slicePos)...)
					}
				}
			}

		// If is any other type
		default:
			// We remove excludePOST fields
			if checkIfTagExists("excludePUT", model.Field(i).Tag.Get("erudito")) {
				source.Field(i).Set(reflect.Zero(model.Field(i).Type))
			} else {
				// If not, whe validate the field (if the model has the function)
				if hasValidateField {
					validationErrors = append(validationErrors, validateField(model.Field(i), source.Field(i), source.Addr(), stack, slicePos)...)
				}
			}
		}
	}

	return validationErrors
}
