package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

func PostHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "POST")

		if maestro.beforeRequestCallback != nil {
			beforeErrors := maestro.beforeRequestCallback(r)
			if beforeErrors != nil {
				SendError(w, 403, beforeErrors)
				return
			}
		}

		modelType := reflect.ValueOf(model).Type()

		/*
		 * DB Connection
		 */
		db := maestro.dBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		/*
		 * JSON Unmarshal
		 */
		modelPostValue := reflect.New(modelType)
		modelPost := modelPostValue.Interface()
		if err := json.NewDecoder(r.Body).Decode(modelPost); err != nil {
			SendSingleError(w, http.StatusUnprocessableEntity, "Given request body was invalid, or some field is in wrong type: "+err.Error(), "")
			return
		}

		/*
		 * Recursive Validation and Cleaning
		 */
		validationErrors := validateAndClearPOST(modelType, modelPostValue, db, r, []string{}, nil)

		if len(validationErrors) > 0 {
			SendError(w, http.StatusForbidden, validationErrors)
			return
		}

		/*
		 * Insert in DB
		 */
		if err := db.Create(modelPost).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "Cannot create a new record - "+err.Error(), "ENTITY_CREATE_ERROR")
			return
		}

		/*
		 * Removal of the ExcludeGET fields
		 */
		for i := 0; i < modelType.NumField(); i++ {
			if checkIfTagExists("excludeGET", modelType.Field(i).Tag.Get("erudito")) {
				reflect.ValueOf(modelPost).Elem().Field(i).Set(reflect.Zero(modelType.Field(i).Type))
			}
		}

		SendData(w, http.StatusCreated, MakeSingularDataStruct(modelType, modelPost))
	})
}

func validateAndClearPOST(model reflect.Type, source reflect.Value, db *gorm.DB, r *http.Request, stack []string, slicePos *uint) []JSendErrorDescription {
	// Final array of errors
	validationErrors := []JSendErrorDescription{}

	// Iterate through Model Fields
	for i := 0; i < model.NumField(); i++ {
		// If have a Package Path, go to next
		if len(model.Field(i).PkgPath) != 0 {
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
				validationErrors = append(validationErrors, validateAndClearPOST(model.Field(i).Type, source.Field(i), db, r, append(stack, model.Field(i).Tag.Get("json")), nil)...)
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
					validationErrors = append(validationErrors, validateAndClearPOST(source.Field(i).Index(j).Type(), source.Field(i).Index(j), db, r, append(stack, model.Field(i).Tag.Get("json")), &pos)...)
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
			if checkIfTagExists("excludePOST", model.Field(i).Tag.Get("erudito")) {
				source.Field(i).Set(reflect.Zero(model.Field(i).Type))
			} else {
				// If not, whe validate the field (if the model has the function)
				if hasValidateField {
					validationErrors = append(validationErrors, validateField(model.Field(i), source.Field(i), source.Addr(), stack, slicePos)...)
				}
			}
		}
	}

	// Call BeforePOST in recursion
	_, hasBeforePOST := model.MethodByName("BeforePOST")
	if hasBeforePOST {
		beforePOSTr := source.MethodByName("BeforePOST").Call([]reflect.Value{
			reflect.ValueOf(db),
			reflect.ValueOf(r),
			source.Addr(),
		})

		if errs := beforePOSTr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
			validationErrors = append(validationErrors, errs...)
		}
	}

	return validationErrors
}

func validateField(modelField reflect.StructField, sourceField reflect.Value, source reflect.Value, stack []string, slicePos *uint) []JSendErrorDescription {
	validateFieldr := source.MethodByName("ValidateField").Call([]reflect.Value{
		reflect.ValueOf(getJSONFieldNameByTag(modelField.Tag.Get("json"))),
		sourceField,
		source,
	})

	if errs := validateFieldr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
		for j := 0; j < len(errs); j++ {
			if len(stack) > 0 {
				refer := strings.Join(stack, ".")
				errs[j].Refer = &refer
			}

			errs[j].Pos = slicePos
		}

		return errs
	}

	return []JSendErrorDescription{}
}
