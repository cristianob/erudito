package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cristianob/erudito/nulls"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

func generatePostModel(w http.ResponseWriter, r *http.Request, db *gorm.DB, modelType reflect.Type, modelS modelStructure, modelUnmarshal map[string]interface{}, maestro *maestro, metaData MiddlewareMetaData, middlewareType int, root bool) (interface{}, MiddlewareMetaData, *[]JSendErrorDescription) {
	rtrModel := reflect.New(modelType)

	for key, value := range modelUnmarshal {
		if key == "id" && value != nil && !root {
			var uintVar uint
			rtrModel.Elem().FieldByName("ID").Set(reflect.ValueOf(value).Convert(reflect.TypeOf(uintVar)))
		}

		field := modelS.getFieldByJson(key)
		if field == nil {
			continue
		}

		// Check for a invalid null
		if value == nil && !field.Nullable {
			return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_CANNOT_NULL", Message: "Field " + field.JsonName + " cannot be null"}}
		}

		modelField := rtrModel.Elem().FieldByName(field.Name)

		if value == nil {
			modelField.Set(reflect.Zero(modelField.Type()))
		} else if field.Type == FIELD_TYPE_COMMON_JSON {
			marshal, _ := json.Marshal(value)

			if field.Nullable {
				newField := reflect.New(modelField.Type())
				newField.MethodByName("Scan").Call([]reflect.Value{reflect.ValueOf(string(marshal))})
				modelField.Set(newField.Elem())
			} else {
				modelField.Set(reflect.ValueOf(string(marshal)))
			}
		} else if field.Type == FIELD_TYPE_RELATION_MODEL {
			if !reflect.TypeOf(value).AssignableTo(reflect.TypeOf(map[string]interface{}{})) {
				return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it needs to be a MODEL"}}
			}

			var relationalModel interface{}
			var err *[]JSendErrorDescription
			relationalModel, metaData, err = generatePostModel(w, r, db, modelField.Type().Elem(), maestro.getModelStructureByName(field.RelationalModel), value.(map[string]interface{}), maestro, metaData, middlewareType, false)
			if err != nil {
				return nil, nil, err
			}

			modelField.Set(reflect.ValueOf(relationalModel))
		} else if field.Type == FIELD_TYPE_RELATION_COL_MODELS || field.Type == FIELD_TYPE_RELATION_COL_MODELS_UPDATE || field.Type == FIELD_TYPE_RELATION_COL_MODELS_REPLACE {
			baseModelField := modelField
			if !baseModelField.IsValid() {
				baseModelField = rtrModel.Elem().FieldByName(field.BaseField)
			}

			relationalArray := reflect.New(baseModelField.Type()).Elem()

			if reflect.TypeOf(value) != reflect.TypeOf([]interface{}{}) {
				return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it needs to be a model ARRAY"}}
			}

			for _, mapIndex := range value.([]interface{}) {
				if !reflect.TypeOf(mapIndex).AssignableTo(reflect.TypeOf(map[string]interface{}{})) {
					return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it needs to be a MODEL array"}}
				}

				var relationalModel interface{}
				var err *[]JSendErrorDescription
				relationalModel, metaData, err = generatePostModel(w, r, db, baseModelField.Type().Elem(), maestro.getModelStructureByName(field.RelationalModel), mapIndex.(map[string]interface{}), maestro, metaData, middlewareType, false)
				if err != nil {
					return nil, nil, err
				}

				relationalArray = reflect.Append(relationalArray, reflect.ValueOf(relationalModel).Elem())
			}

			baseModelField.Set(relationalArray)
		} else if field.Type == FIELD_TYPE_COMMON_TIME {
			str := value.(string)

			t, err := utilsTryParseISOTime(str)
			if err != nil {
				return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type. Date can be YYYY-MM-DDTHH:MM:SS.mmmmmmmmmZ, YYYY-MM-DDTHH:MM:SS.mmmmmm-TT:hh, YYYY-MM-DDTHH:MM:SSZ or YYYY-MM-DD"}}
			}

			modelField.Set(reflect.ValueOf(t))

		} else if field.Type == FIELD_TYPE_RELATION_ID || field.Type == FIELD_TYPE_COMMON || field.Type == FIELD_TYPE_INTERNAL {
			if field.Nullable {
				if modelField.Type() == reflect.TypeOf(nulls.Time{}) {
					str := value.(string)

					t, err := utilsTryParseISOTime(str)
					if err != nil {
						return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type. Date can be YYYY-MM-DDTHH:MM:SS.mmmmmmmmmZ, YYYY-MM-DDTHH:MM:SS.mmmmmm-TT:hh, YYYY-MM-DDTHH:MM:SSZ or YYYY-MM-DD"}}
					}

					newField := reflect.New(modelField.Type())
					newField.MethodByName("Scan").Call([]reflect.Value{reflect.ValueOf(t)})
					modelField.Set(newField.Elem())
				} else {
					newField := reflect.New(modelField.Type())
					newField.MethodByName("Scan").Call([]reflect.Value{reflect.ValueOf(value)})
					modelField.Set(newField.Elem())
				}
			} else {
				if !reflect.ValueOf(value).Type().ConvertibleTo(modelField.Type()) {
					return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it should be " + modelField.Type().Name()}}
				}

				modelField.Set(reflect.ValueOf(value).Convert(modelField.Type()))
			}
		}
	}

	response := utilsRunMiddlewareBefore(w, r, db, rtrModel, maestro, metaData, middlewareType, root)
	if response.Error != nil {
		return nil, nil, response.Error
	}

	return response.Model.Interface(), response.Meta, nil
}

func generateReturnModel(w http.ResponseWriter, r *http.Request, db *gorm.DB, modelType reflect.Type, modelS modelStructure, modelResponse reflect.Value, maestro *maestro, metaData MiddlewareMetaData, middlewareType int, root bool) (interface{}, MiddlewareMetaData, *[]JSendErrorDescription) {
	if modelResponse.Type().Kind() == reflect.Ptr {
		modelResponse = modelResponse.Elem()
	}

	mapResponse := structs.Map(modelResponse.Interface())
	mapReturn := map[string]interface{}{}

	if fullModel, ok := mapResponse["FullModel"]; ok {
		mapReturn["id"] = fullModel.(map[string]interface{})["ID"]
		mapReturn["created_at"] = fullModel.(map[string]interface{})["CreatedAt"]
		mapReturn["updated_at"] = fullModel.(map[string]interface{})["UpdatedAt"]
		mapReturn["deleted_at"] = fullModel.(map[string]interface{})["DeletedAt"]
	}

	if hardDeleteModel, ok := mapResponse["HardDeleteModel"]; ok {
		mapReturn["id"] = hardDeleteModel.(map[string]interface{})["ID"]
		mapReturn["created_at"] = hardDeleteModel.(map[string]interface{})["CreatedAt"]
		mapReturn["updated_at"] = hardDeleteModel.(map[string]interface{})["UpdatedAt"]
	}

	if simpleModel, ok := mapResponse["SimpleModel"]; ok {
		mapReturn["id"] = simpleModel.(map[string]interface{})["ID"]
	}

	for _, field := range modelS.Fields {
		if fieldValue, ok := mapResponse[field.Name]; ok {
			if field.Type == FIELD_TYPE_COMMON_TIME {
				mapReturn[field.JsonName] = fieldValue
			} else if field.Type == FIELD_TYPE_COMMON_JSON {
				if field.Nullable {
					var unmarshal interface{}
					modelValue := modelResponse.FieldByName(field.Name).Interface().(nulls.String)
					if modelValue.Valid {
						json.Unmarshal([]byte(modelValue.String), &unmarshal)
						mapReturn[field.JsonName] = unmarshal
					} else {
						mapReturn[field.JsonName] = nil
					}
				} else {
					var unmarshal interface{}
					json.Unmarshal([]byte(fieldValue.(string)), &unmarshal)
					mapReturn[field.JsonName] = unmarshal
				}
			} else if field.Type == FIELD_TYPE_COMMON || field.Type == FIELD_TYPE_INTERNAL || field.Type == FIELD_TYPE_RELATION_ID {
				if field.Nullable {
					fieldValueR := modelResponse.FieldByName(field.Name)
					if fieldValueR.FieldByName("Valid").Interface().(bool) {
						mapReturn[field.JsonName] = fieldValueR.Field(0).Interface()
					} else {
						mapReturn[field.JsonName] = nil
					}
				} else {
					mapReturn[field.JsonName] = fieldValue
				}
			} else if field.Type == FIELD_TYPE_RELATION_COL_MODELS {
				relationalMap := []interface{}{}

				fieldStructValue := modelResponse.FieldByName(field.Name)

				if fieldStructValue.Pointer() == 0 {
					continue
				}

				for i := 0; i < fieldStructValue.Len(); i++ {
					fieldV := fieldStructValue.Index(i)

					var relationalRtr interface{}
					relationalRtr, metaData, _ = generateReturnModel(w, r, db, modelType, maestro.getModelStructureByName(field.RelationalModel), fieldV, maestro, metaData, middlewareType, false)
					relationalMap = append(relationalMap, relationalRtr)
				}

				mapReturn[field.JsonName] = relationalMap
			} else if field.Type == FIELD_TYPE_RELATION_MODEL {
				var relationalRtr interface{}

				fieldStructValue := modelResponse.FieldByName(field.Name)

				if fieldStructValue.IsNil() {
					continue
				}

				relationalRtr, metaData, _ = generateReturnModel(w, r, db, modelType, maestro.getModelStructureByName(field.RelationalModel), fieldStructValue, maestro, metaData, middlewareType, false)
				mapReturn[field.JsonName] = relationalRtr
			}
		}
	}

	response := utilsRunMiddlewareAfter(w, r, db, modelResponse, mapReturn, maestro, metaData, root)
	if response.Error != nil {
		return nil, metaData, response.Error
	}

	return response.Response, metaData, nil
}

func insertMultipleRelations(w http.ResponseWriter, r *http.Request, db *gorm.DB, modelType reflect.Type, modelS modelStructure, modelUnmarshal map[string]interface{}, modelResponse reflect.Value, maestro *maestro, root bool) (interface{}, *[]JSendErrorDescription) {
	modelResponseElem := modelResponse
	if modelResponse.Type().Kind() == reflect.Ptr {
		modelResponseElem = modelResponse.Elem()
	}

	mapResponse := structs.Map(modelResponse.Interface())

	var modelID uint = 0
	if fullModel, ok := mapResponse["FullModel"]; ok {
		modelID = fullModel.(map[string]interface{})["ID"].(uint)
	}

	if hardDeleteModel, ok := mapResponse["HardDeleteModel"]; ok {
		modelID = hardDeleteModel.(map[string]interface{})["ID"].(uint)
	}

	if simpleModel, ok := mapResponse["SimpleModel"]; ok {
		modelID = simpleModel.(map[string]interface{})["ID"].(uint)
	}

	model1DB := reflect.New(reflect.TypeOf(modelS.Model)).Interface()
	if notFound := db.First(model1DB, modelID).RecordNotFound(); notFound {
		return nil, &[]JSendErrorDescription{{Code: "POSTED_ID_DONT_EXIST", Message: "This is an Erudito's error, please report"}}
	}

	for key, value := range modelUnmarshal {
		field := modelS.getFieldByJson(key)

		if field == nil {
			continue
		}

		modelField := modelResponseElem.FieldByName(field.Name)
		if !modelField.IsValid() {
			modelField = modelResponseElem.FieldByName(field.BaseField)
		}

		if field.Type == FIELD_TYPE_RELATION_COL_MODELS || field.Type == FIELD_TYPE_RELATION_COL_MODELS_UPDATE || field.Type == FIELD_TYPE_RELATION_COL_MODELS_REPLACE {
			var err error

			switch field.Type {
			case FIELD_TYPE_RELATION_COL_MODELS_UPDATE:
				err = db.Model(model1DB).Association(field.BaseField).Replace(modelField.Interface()).Error
				if err != nil {
					return nil, &[]JSendErrorDescription{{Code: "RELATION_ERROR", Message: "Relation update error: " + err.Error()}}
				}

			case FIELD_TYPE_RELATION_COL_MODELS_REPLACE:
				relationalModel := maestro.getModelStructureByName(field.RelationalModel)
				model2DBCheck := reflect.New(reflect.SliceOf(reflect.TypeOf(relationalModel.Model))).Interface()

				db.Model(model1DB).Association(field.BaseField).Find(model2DBCheck)

				model2DBCheckValue := reflect.ValueOf(model2DBCheck).Elem()
				for i := 0; i < model2DBCheckValue.Len(); i++ {
					exists := false
					for j := 0; j < modelField.Len(); j++ {
						if model2DBCheckValue.Index(i).FieldByName("ID").Interface().(uint) == modelField.Index(j).FieldByName("ID").Interface().(uint) {
							exists = true
						}
					}

					if !exists {
						db.Delete(model2DBCheckValue.Index(i).Interface())
					}
				}

				err = db.Model(model1DB).Association(field.BaseField).Replace(modelField.Interface()).Error
				if err != nil {
					return nil, &[]JSendErrorDescription{{Code: "RELATION_ERROR", Message: "Relation update error: " + err.Error()}}
				}
			}

			relationalArray := reflect.New(modelField.Type()).Elem()

			for i, mapIndex := range value.([]interface{}) {
				var relationalModel interface{}
				var err *[]JSendErrorDescription
				relationalModel, err = insertMultipleRelations(w, r, db, modelField.Type().Elem(), maestro.getModelStructureByName(field.RelationalModel), mapIndex.(map[string]interface{}), modelField.Index(i), maestro, false)
				if err != nil {
					return nil, err
				}

				relationalArray = reflect.Append(relationalArray, reflect.ValueOf(relationalModel))
			}

			modelField.Set(relationalArray)
		} else if field.Type == FIELD_TYPE_RELATION_COL_IDS || field.Type == FIELD_TYPE_RELATION_COL_IDS_UPDATE || field.Type == FIELD_TYPE_RELATION_COL_IDS_REPLACE {
			relationalModel := maestro.getModelStructureByName(field.RelationalModel)

			model2DBs := []interface{}{}
			for _, relIDInterface := range value.([]interface{}) {
				if reflect.TypeOf(relIDInterface).Kind() != reflect.Float64 {
					return nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it should be a ID array"}}
				}

				relID := int(relIDInterface.(float64))
				model2DB := reflect.New(reflect.TypeOf(relationalModel.Model)).Interface()

				if notFound := db.First(model2DB, relID).RecordNotFound(); notFound {
					return nil, &[]JSendErrorDescription{{Code: "RELATION_ID_DONT_EXISTS", Message: "Relation " + strconv.FormatInt(int64(relID), 10) + " of field " + field.Name + " doesn't exists"}}
				}

				model2DBs = append(model2DBs, reflect.ValueOf(model2DB).Elem().Interface())
			}

			var err error

			switch field.Type {
			case FIELD_TYPE_RELATION_COL_IDS:
				err = db.Model(model1DB).Association(field.BaseField).Append(model2DBs...).Error
			case FIELD_TYPE_RELATION_COL_IDS_UPDATE:
				err = db.Model(model1DB).Association(field.BaseField).Replace(model2DBs...).Error
			case FIELD_TYPE_RELATION_COL_IDS_REPLACE:
				model2DBCheck := reflect.New(reflect.SliceOf(reflect.TypeOf(relationalModel.Model))).Interface()
				db.Model(model1DB).Association(field.BaseField).Find(model2DBCheck)

				model2DBCheckValue := reflect.ValueOf(model2DBCheck).Elem()
				for i := 0; i < model2DBCheckValue.Len(); i++ {
					exists := false
					for _, m := range model2DBs {
						model2DBValue := reflect.ValueOf(m)
						if model2DBCheckValue.Index(i).FieldByName("ID").Interface().(uint) == model2DBValue.FieldByName("ID").Interface().(uint) {
							exists = true
						}
					}

					if !exists {
						db.Delete(model2DBCheckValue.Index(i).Interface())
					}
				}

				err = db.Model(model1DB).Association(field.BaseField).Replace(model2DBs...).Error
			}

			if err != nil {
				return nil, &[]JSendErrorDescription{{Code: "RELATION_ERROR", Message: "Relation update error: " + err.Error()}}
			}

			modelResponseElem.FieldByName(field.BaseField).Set(reflect.AppendSlice(modelResponseElem.FieldByName(field.BaseField), reflect.ValueOf(model1DB).Elem().FieldByName(field.BaseField)))
		}
	}

	return modelResponse.Interface(), nil
}

func utilsRunMiddlewaresInitial(middlewaresInitial []MiddlewareInitial, w http.ResponseWriter, r *http.Request, maestro *maestro, metaData MiddlewareMetaData, middlewareType int) MiddlewareInitialReturn {
	mwReturn := MiddlewareInitialReturn{
		R:     r,
		W:     w,
		Meta:  metaData,
		Error: nil,
	}

	for _, middleware := range middlewaresInitial {
		if middleware.Type != MIDDLEWARE_TYPE_GLOBAL && middleware.Type != middlewareType {
			continue
		}

		mwReturn = middleware.Function(MiddlewareInitialData{
			R:    mwReturn.R,
			W:    mwReturn.W,
			Meta: mwReturn.Meta,
		})

		if mwReturn.Error != nil {
			return mwReturn
		}
	}

	return mwReturn
}

func utilsRunMiddlewareBefore(w http.ResponseWriter, r *http.Request, db *gorm.DB, model reflect.Value, maestro *maestro, metaData MiddlewareMetaData, middlewareType int, root bool) MiddlewareBeforeReturn {
	modelMiddlewaresValue := model.MethodByName("MiddlewareBefore").Call([]reflect.Value{})
	modelMiddlewares := modelMiddlewaresValue[0].Interface().([]MiddlewareBefore)

	mpReturn := MiddlewareBeforeReturn{
		R:     r,
		W:     w,
		Meta:  metaData,
		Model: model,
		Error: nil,
	}

	for _, middleware := range modelMiddlewares {
		if !root && middleware.Level == MIDDLEWARE_LEVEL_ROOT {
			continue
		}

		if middleware.Type != MIDDLEWARE_TYPE_GLOBAL && middleware.Type != middlewareType {
			continue
		}

		mpReturn = middleware.Function(MiddlewareBeforeData{
			R:      mpReturn.R,
			W:      mpReturn.W,
			DbConn: db,
			Meta:   mpReturn.Meta,
			Model:  mpReturn.Model,
		})

		if mpReturn.Error != nil {
			return mpReturn
		}
	}

	return mpReturn
}

func utilsRunMiddlewareAfter(w http.ResponseWriter, r *http.Request, db *gorm.DB, modelResponse reflect.Value, model map[string]interface{}, maestro *maestro, metaData MiddlewareMetaData, root bool) MiddlewareAfterReturn {
	modelValue := modelResponse.Interface().(Model)
	modelMiddlewares := modelValue.MiddlewareAfter()

	mpReturn := MiddlewareAfterReturn{
		R:        r,
		W:        w,
		Meta:     metaData,
		Response: model,
		Error:    nil,
	}

	for _, middleware := range modelMiddlewares {
		if !root && middleware.Level == MIDDLEWARE_LEVEL_ROOT {
			continue
		}

		if middleware.Type != MIDDLEWARE_TYPE_GLOBAL && middleware.Type != MIDDLEWARE_TYPE_POST {
			continue
		}

		mpReturn = middleware.Function(MiddlewareAfterData{
			R:        mpReturn.R,
			W:        mpReturn.W,
			DbConn:   db,
			Meta:     mpReturn.Meta,
			Response: mpReturn.Response,
		})

		if mpReturn.Error != nil {
			return mpReturn
		}
	}

	return mpReturn
}

func utilsTryParseISOTime(str string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.000000000Z"
	t, err := time.Parse(layout, str)

	if err != nil {
		layout := "2006-01-02T15:04:05.999999-07:00"
		t, err = time.Parse(layout, str)
	}

	if err != nil {
		layout := "2006-01-02T15:04:05Z"
		t, err = time.Parse(layout, str)
	}

	if err != nil {
		layout := "2006-01-02"
		t, err = time.Parse(layout, str)
	}

	return t, err
}

func getDBWithSearchCriterias(db *gorm.DB, r *http.Request, modelNew interface{}, modelType reflect.Type) *gorm.DB {
	unscoped := false

	softDeletedResponse, softDeletedOK := r.URL.Query()["del"]
	if softDeletedOK {
		if softDeletedResponse[0] == "true" {
			unscoped = true
		}
	}

	if unscoped {
		db = db.Unscoped()
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
			for _, getFieldItem := range getField {
				getFields := strings.Split(getFieldItem, "|")

				for gi, gf := range getFields {
					switch modelType.Field(i).Type.Kind() {
					case reflect.String:
						if gi == 0 {
							db = db.Where(fieldJSON+" LIKE ?", gf)
						} else {
							db = db.Or(fieldJSON+" LIKE ?", gf)
						}

					default:
						if gi == 0 {
							db = db.Where(fieldJSON+" = ?", gf)
						} else {
							db = db.Or(fieldJSON+" = ?", gf)
						}
					}
				}
			}
		}

		if getField, ok := r.URL.Query()[fieldJSON+"_egt"]; ok {
			db = db.Where(fieldJSON+" >= ?", getField)
		}

		if getField, ok := r.URL.Query()[fieldJSON+"_elt"]; ok {
			db = db.Where(fieldJSON+" <= ?", getField)
		}
	}

	if getField, ok := r.URL.Query()["id"]; ok {
		for _, getFieldItem := range getField {
			getFields := strings.Split(getFieldItem, "|")

			for gi, gf := range getFields {
				if gi == 0 {
					db = db.Where("id = ?", gf)
				} else {
					db = db.Or("id = ?", gf)
				}
			}
		}
	}

	if getField, ok := r.URL.Query()["created_at_egt"]; ok {
		db = db.Where("created_at >= ?", getField)
	}

	if getField, ok := r.URL.Query()["created_at_elt"]; ok {
		db = db.Where("created_at <= ?", getField)
	}

	relString, ok := r.URL.Query()["rel"]
	if ok {
		rels := strings.Split(relString[0], ",")

		for _, rel := range rels {
			db = db.Preload(upperCamelCase(rel), func(dbPreload *gorm.DB) *gorm.DB {
				if unscoped {
					dbPreload = dbPreload.Unscoped()
				}

				return dbPreload
			})
		}
	}

	return db
}
