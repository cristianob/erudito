package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
)

func PostHandler(model Model, DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		modelType := reflect.ValueOf(model).Type()

		/*
		 * DB Connection
		 */
		db := DBPoolCallback(r)
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
		validationErrors := validateAndClear(modelType, modelPostValue, db, r)

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

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "POST")
		w.Header().Add("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization")
		w.Header().Add("Access-Control-Max-Age", "1728000")

		SendData(w, http.StatusCreated, MakeSingularDataStruct(modelType, modelPost))
	})
}

func validateAndClear(model reflect.Type, source reflect.Value, db *gorm.DB, r *http.Request) []JSendErrorDescription {
	validationErrors := []JSendErrorDescription{}

	for i := 0; i < model.NumField(); i++ {
		if len(model.Field(i).PkgPath) != 0 {
			continue
		}

		if source.Type().Kind() == reflect.Ptr {
			source = source.Elem()
		}

		_, hasValidateField := model.MethodByName("ValidateField")

		switch model.Field(i).Type.Kind() {
		case reflect.Struct:
			if model.Field(i).Type.Implements(reflect.TypeOf((*Model)(nil)).Elem()) {
				validationErrors = append(validationErrors, validateAndClear(model.Field(i).Type, source.Field(i), db, r)...)
			} else {
				if hasValidateField {
					beforePOSTr := source.MethodByName("ValidateField").Call([]reflect.Value{
						reflect.ValueOf(model.Field(i).Tag.Get("json")),
						source.Field(i),
						source.Addr(),
					})

					if errs := beforePOSTr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
						validationErrors = append(validationErrors, errs...)
					}
				}
			}

		case reflect.Slice:
			for j := 0; j < source.Field(i).Len(); j++ {
				validationErrors = append(validationErrors, validateAndClear(reflect.TypeOf(source.Field(i).Index(j)), source.Field(i).Index(j), db, r)...)
			}

		default:
			if checkIfTagExists("excludePOST", model.Field(i).Tag.Get("erudito")) {
				source.Field(i).Set(reflect.Zero(model.Field(i).Type))
			} else {
				if hasValidateField {
					beforePOSTr := source.MethodByName("ValidateField").Call([]reflect.Value{
						reflect.ValueOf(model.Field(i).Tag.Get("json")),
						source.Field(i),
						source.Addr(),
					})

					if errs := beforePOSTr[0].Interface().([]JSendErrorDescription); errs != nil && len(errs) > 0 {
						validationErrors = append(validationErrors, errs...)
					}
				}
			}
		}
	}

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
