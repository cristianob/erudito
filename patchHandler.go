package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func PatchHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "PATCH")

		if maestro.beforeRequestCallback != nil {
			beforeErrors := maestro.beforeRequestCallback(r)
			if beforeErrors != nil {
				SendError(w, 403, beforeErrors)
				return
			}
		}

		modelType := reflect.ValueOf(model).Type()
		modelNewValue := reflect.New(modelType)

		modelDBValue := reflect.New(modelType)
		modelDB := modelDBValue.Interface()

		var modelSent map[string]interface{}

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
		 * Getting DB model
		 */
		if notFound := db.First(modelDB, ModelIDField).RecordNotFound(); notFound {
			SendSingleError(w, http.StatusForbidden, "Entity desn't exists", "ENTITY_DONT_EXISTS")
			return
		}

		/*
		 * JSON Unmarshal
		 */
		if err := json.NewDecoder(r.Body).Decode(&modelSent); err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type.", "")
			return
		}

		allErrs := []JSendErrorDescription{}
		modelUpdate := map[string]interface{}{}
		for i := 0; i < modelType.NumField(); i++ {
			jsonFieldName := getJSONFieldNameByTag(modelType.Field(i).Tag.Get("json"))

			// Verify if is a internal field
			if jsonFieldName == "" || jsonFieldName == "-" {
				continue
			}

			// Verify if exists in the Model
			if _, exists := modelSent[jsonFieldName]; !exists {
				continue
			}

			// Verify if is a Erudito Model
			if fieldIsAEruditoModel(modelType.Field(i)) {
				allErrs = append(allErrs, JSendErrorDescription{
					Code:    "INVALID_ASSOCIATION",
					Message: "Field '" + jsonFieldName + "' cannot be send in a PATCH request",
				})

				continue
			}

			// If is a ponter in model, turn to a pointer
			if modelType.Field(i).Type.AssignableTo(reflect.PtrTo(reflect.TypeOf(modelSent[jsonFieldName]))) {
				newValue := reflect.New(reflect.TypeOf(modelSent[jsonFieldName]))
				newValue.Elem().Set(reflect.ValueOf(modelSent[jsonFieldName]))
				modelSent[jsonFieldName] = newValue.Interface()
			}

			// Verify if is the same type or a pointer of the type
			// sameType := reflect.TypeOf(modelSent[jsonFieldName]).AssignableTo(modelType.Field(i).Type)
			// if !sameType {
			// 	allErrs = append(allErrs, JSendErrorDescription{
			// 		Code:    "INVALID_TYPE",
			// 		Message: "Field '" + jsonFieldName + "' needs to be a " + modelType.Field(i).Type.String(),
			// 	})

			// 	continue
			// }

			// Validating fields
			_, hasValidateField := modelType.MethodByName("ValidateField")
			if hasValidateField {
				validateFieldr := modelNewValue.MethodByName("ValidateField").Call([]reflect.Value{
					reflect.ValueOf(jsonFieldName),
					reflect.ValueOf(modelSent[jsonFieldName]),
					modelDBValue,
				})

				if errs := validateFieldr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
					allErrs = append(allErrs, errs...)
				}
			}

			modelUpdate[jsonFieldName] = modelSent[jsonFieldName]
		}

		if len(allErrs) > 0 {
			SendError(w, http.StatusForbidden, allErrs)
			return
		}

		/*
		 * Saving the new model
		 */
		if err := db.Model(modelDB).Set("gorm:save_associations", false).Set("gorm:association_save_reference", false).Updates(modelSent).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(modelType, modelDB))
	})
}
