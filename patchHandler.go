package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/cristianob/erudito/nulls"
)

func PatchHandler(modelZero Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "PATCH")

		modelType := reflect.ValueOf(modelZero).Type()
		modelS := maestro.getModelStructure(modelZero)

		modelDBValue := reflect.New(modelType)
		modelDB := modelDBValue.Interface()

		var modelSent map[string]interface{}

		/*
		 * Middleware Initial
		 */
		metaData := MiddlewareMetaData{}

		mwInitial := utilsRunMiddlewaresInitial(maestro.MiddlewaresInitial, w, r, maestro, metaData, MIDDLEWARE_TYPE_PATCH)
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
			if reflect.TypeOf(modelSent[jsonFieldName]) == nil {
				modelSent[jsonFieldName] = reflect.New(modelType.Field(i).Type).Elem().Interface()
			} else if modelType.Field(i).Type.AssignableTo(reflect.PtrTo(reflect.TypeOf(modelSent[jsonFieldName]))) {
				newValue := reflect.New(reflect.TypeOf(modelSent[jsonFieldName]))
				newValue.Elem().Set(reflect.ValueOf(modelSent[jsonFieldName]))
				modelSent[jsonFieldName] = newValue.Interface()
			}

			// If is a time field, we convert to SQL format
			if modelType.Field(i).Type.AssignableTo(reflect.TypeOf(nulls.Time{})) ||
				modelType.Field(i).Type.AssignableTo(reflect.TypeOf(time.Time{})) {

				if reflect.TypeOf(modelSent[jsonFieldName]).Kind() != reflect.String {
					allErrs = append(allErrs, JSendErrorDescription{
						Code:    "INVALID_TIME",
						Message: "Field '" + jsonFieldName + "' is not a valid ISO time",
					})

					continue
				}

				time, err := time.Parse(time.RFC3339, modelSent[jsonFieldName].(string))
				if err != nil {
					allErrs = append(allErrs, JSendErrorDescription{
						Code:    "INVALID_TIME",
						Message: "Field '" + jsonFieldName + "' is not a valid ISO time: " + err.Error(),
					})

					continue
				}

				modelSent[jsonFieldName] = time
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
		if err := db.Model(modelDB).Set("gorm:save_associations", false).Set("gorm:association_save_reference", false).Updates(modelUpdate).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot update record - "+err.Error(), "ENTITY_UPDATE_ERROR")
			return
		}

		/*
		 * Generate Response
		 */
		modelGenerated, _, errR := generateReturnModel(w, r, db, modelType, modelS, reflect.ValueOf(modelDB), maestro, metaData, MIDDLEWARE_TYPE_PATCH, true)

		if errR != nil {
			SendError(w, http.StatusForbidden, *errR)
			return
		}

		SendData(w, http.StatusAccepted, MakeSingularDataStruct(modelType, modelGenerated, modelS))
	})
}
