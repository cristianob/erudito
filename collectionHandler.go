package erudito

import (
	"net/http"
	"reflect"
	"strings"
)

func CollectionHandler(model Model, maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		if maestro.beforeRequestCallback != nil {
			beforeErrors := maestro.beforeRequestCallback(r)
			if beforeErrors != nil {
				SendError(w, 403, beforeErrors)
				return
			}
		}

		modelType := reflect.ValueOf(model).Type()
		modelNew := reflect.New(modelType).Interface()

		db := maestro.dBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		if order, ok := r.URL.Query()["order"]; ok {
			db = db.Order(order[0])
		}

		if limit, ok := r.URL.Query()["limit"]; ok {
			db = db.Limit(limit[0])
		} else {
			db = db.Limit(5000)
		}

		if offset, ok := r.URL.Query()["offset"]; ok {
			db = db.Offset(offset[0])
		}

		modelNewValue := reflect.ValueOf(modelNew).Elem()
		for i := 0; i < modelNewValue.NumField(); i++ {
			fieldJSON := modelType.Field(i).Tag.Get("json")

			if getField, ok := r.URL.Query()[fieldJSON]; ok {
				switch modelType.Field(i).Type.Kind() {
				case reflect.String:
					db = db.Where(fieldJSON+" LIKE ?", getField)

				case reflect.Int, reflect.Uint:
					db = db.Where(fieldJSON+" = ?", getField)
				}
			}

			if getField, ok := r.URL.Query()[fieldJSON+"_egt"]; ok {
				db = db.Where(fieldJSON+" >= ?", getField)
			}

			if getField, ok := r.URL.Query()[fieldJSON+"_elt"]; ok {
				db = db.Where(fieldJSON+" <= ?", getField)
			}
		}

		relString, ok := r.URL.Query()["rel"]
		if ok {
			rels := strings.Split(relString[0], ",")

			for _, rel := range rels {
				db = db.Preload(upperCamelCase(rel))
			}
		}

		softDeleted, ok := r.URL.Query()["del"]
		if ok {
			if softDeleted[0] == "true" {
				db = db.Unscoped()
			}
		}

		modelSlice := reflect.New(reflect.SliceOf(modelType))
		if err := db.Find(modelSlice.Interface()).Error; err != nil {
			SendSingleError(w, http.StatusForbidden, "There is an error in your query: "+err.Error(), "QUERY_ERROR")
			return
		}

		/*
		 * Removal of the ExcludeGET fields
		 */
		modelSliceRemoved := reflect.New(reflect.SliceOf(modelType)).Elem()
		_, ok = reflect.TypeOf(model).MethodByName("BeforeCollectionResponse")
		if ok {
			for el := 0; el < modelSlice.Elem().Len(); el++ {
				beforeCollectionResponse := modelSlice.Elem().Index(el).MethodByName("BeforeCollectionResponse").Call([]reflect.Value{
					reflect.ValueOf(maestro.dBPoolCallback(r)),
					reflect.ValueOf(r),
					modelSlice.Elem().Index(el).Addr(),
				})

				if beforeCollectionResponse[0].Interface().(bool) {
					modelSliceRemoved.Set(reflect.Append(modelSliceRemoved, modelSlice.Elem().Index(el)))
				}
			}
		} else {
			modelSliceRemoved.Set(modelSlice.Elem())
		}

		for el := 0; el < modelSliceRemoved.Len(); el++ {
			for i := 0; i < modelType.NumField(); i++ {
				if checkIfTagExists("excludeGET", modelType.Field(i).Tag.Get("erudito")) {
					modelSliceRemoved.Index(el).Field(i).Set(reflect.Zero(modelType.Field(i).Type))
				}
			}
		}

		SendData(w, http.StatusOK, MakeArrayDataStruct(modelType, modelSliceRemoved))
	})
}
