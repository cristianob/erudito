package erudito

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cristianob/erudito/nulls"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

func generatePostModel(w http.ResponseWriter, r *http.Request, db *gorm.DB, modelType reflect.Type, modelS modelStructure, modelUnmarshal map[string]interface{}, maestro *maestro, metaData MidlewareMetaData, root bool) (interface{}, MidlewareMetaData, *[]JSendErrorDescription) {
	rtrModel := reflect.New(modelType)

	for key, value := range modelUnmarshal {
		if key == "id" && !root {
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
			relationalModel, metaData, err = generatePostModel(w, r, db, modelField.Type().Elem(), maestro.getModelStructureByName(field.RelationalModel), value.(map[string]interface{}), maestro, metaData, false)
			if err != nil {
				return nil, nil, err
			}

			modelField.Set(reflect.ValueOf(relationalModel))
		} else if field.Type == FIELD_TYPE_RELATION_COL_MODELS {
			relationalArray := reflect.New(modelField.Type()).Elem()

			if reflect.TypeOf(value) != reflect.TypeOf([]interface{}{}) {
				return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it needs to be a model ARRAY"}}
			}

			for _, mapIndex := range value.([]interface{}) {
				if !reflect.TypeOf(mapIndex).AssignableTo(reflect.TypeOf(map[string]interface{}{})) {
					return nil, nil, &[]JSendErrorDescription{{Code: "FIELD_WRONG_TYPE", Message: "Field " + field.JsonName + " is in a wrong type, it needs to be a MODEL array"}}
				}

				var relationalModel interface{}
				var err *[]JSendErrorDescription
				relationalModel, metaData, err = generatePostModel(w, r, db, modelField.Type().Elem(), maestro.getModelStructureByName(field.RelationalModel), mapIndex.(map[string]interface{}), maestro, metaData, false)
				if err != nil {
					return nil, nil, err
				}

				relationalArray = reflect.Append(relationalArray, reflect.ValueOf(relationalModel).Elem())
			}

			modelField.Set(relationalArray)
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

	response := utilsRunMiddlewaresPRE(w, r, db, rtrModel.Elem(), maestro, metaData, root)
	if response.Error != nil {
		return nil, nil, response.Error
	}

	return rtrModel.Interface(), response.Meta, nil
}

func utilsRunMiddlewaresPRE(w http.ResponseWriter, r *http.Request, db *gorm.DB, model reflect.Value, maestro *maestro, metaData MidlewareMetaData, root bool) MiddlewarePREReturn {
	modelValue := model.Interface().(Model)
	modelMiddlewares := modelValue.MiddlewaresPRE()

	mpReturn := MiddlewarePREReturn{
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

		if middleware.Type != MIDDLEWARE_TYPE_GLOBAL && middleware.Type != MIDDLEWARE_TYPE_POST {
			continue
		}

		mpReturn = middleware.Function(MiddlewarePREData{
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

func generateReturnModel(w http.ResponseWriter, r *http.Request, db *gorm.DB, modelType reflect.Type, modelS modelStructure, modelResponse reflect.Value, maestro *maestro, metaData MidlewareMetaData, root bool) (interface{}, MidlewareMetaData, *[]JSendErrorDescription) {
	if modelResponse.Type().Kind() == reflect.Ptr {
		modelResponse = modelResponse.Elem()
	}

	mapResponse := structs.Map(modelResponse.Interface())
	mapReturn := map[string]interface{}{}

	if fullModel, ok := mapResponse["FullModel"]; ok {
		mapReturn["id"] = fullModel.(map[string]interface{})["ID"]
		mapReturn["created_at"] = fullModel.(map[string]interface{})["CreatedAt"]
		mapReturn["updated_at"] = fullModel.(map[string]interface{})["UpdatedAt"]
	}

	if hardDeleteModel, ok := mapResponse["HardDeleteModel"]; ok {
		mapReturn["id"] = hardDeleteModel.(map[string]interface{})["ID"]
		mapReturn["created_at"] = hardDeleteModel.(map[string]interface{})["CreatedAt"]
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
				for i := 0; i < fieldStructValue.Len(); i++ {
					fieldV := fieldStructValue.Index(i)

					var relationalRtr interface{}
					relationalRtr, metaData, _ = generateReturnModel(w, r, db, modelType, maestro.getModelStructureByName(field.RelationalModel), fieldV, maestro, metaData, false)
					relationalMap = append(relationalMap, relationalRtr)
				}

				mapReturn[field.JsonName] = relationalMap
			} else if field.Type == FIELD_TYPE_RELATION_MODEL {
				var relationalRtr interface{}

				fieldStructValue := modelResponse.FieldByName(field.Name)

				if fieldStructValue.IsNil() {
					continue
				}

				relationalRtr, metaData, _ = generateReturnModel(w, r, db, modelType, maestro.getModelStructureByName(field.RelationalModel), fieldStructValue, maestro, metaData, false)
				mapReturn[field.JsonName] = relationalRtr
			}
		}
	}

	// // response := utilsRunMiddlewaresPRE(w, r, db, rtrModel.Elem(), maestro, metaData, root)
	// // if response.Error != nil {
	// // 	return nil, response.Error
	// // }

	return mapReturn, metaData, nil
}

// func utilsRunMiddlewaresPOS(w http.ResponseWriter, r *http.Request, db *gorm.DB, model reflect.Value, maestro *maestro, metaData MidlewareMetaData, root bool) MiddlewarePREReturn {
// 	modelValue := model.Interface().(Model)
// 	modelMiddlewares := modelValue.MiddlewaresPRE()

// 	mpReturn := MiddlewarePREReturn{
// 		R:     r,
// 		W:     w,
// 		Meta:  metaData,
// 		Model: model,
// 		Error: nil,
// 	}

// 	for _, middleware := range modelMiddlewares {
// 		if !root && middleware.Level == MIDDLEWARE_LEVEL_ROOT {
// 			continue
// 		}

// 		if middleware.Type != MIDDLEWARE_TYPE_GLOBAL && middleware.Type != MIDDLEWARE_TYPE_POST {
// 			continue
// 		}

// 		mpReturn = middleware.Function(MiddlewarePREData{
// 			R:      mpReturn.R,
// 			W:      mpReturn.W,
// 			DbConn: db,
// 			Meta:   mpReturn.Meta,
// 			Model:  mpReturn.Model,
// 		})

// 		if mpReturn.Error != nil {
// 			return mpReturn
// 		}
// 	}

// 	return mpReturn
// }

func insertMultipleRelations(db *gorm.DB, modelType reflect.Type, modelS modelStructure, modelUnmarshal map[string]interface{}, modelResponse reflect.Value, maestro *maestro, root bool) (interface{}, *[]JSendErrorDescription) {
	modelResponseElem := modelResponse
	if modelResponse.Type().Kind() == reflect.Ptr {
		modelResponseElem = modelResponse.Elem()
	}

	log.Println(root, modelResponseElem)

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

	for key, value := range modelUnmarshal {
		field := modelS.getFieldByJson(key)

		modelField := modelResponseElem.FieldByName(field.Name)

		if field.Type == FIELD_TYPE_RELATION_COL_MODELS {
			relationalArray := reflect.New(modelField.Type()).Elem()

			for i, mapIndex := range value.([]interface{}) {
				var relationalModel interface{}
				var err *[]JSendErrorDescription
				relationalModel, err = insertMultipleRelations(db, modelField.Type().Elem(), maestro.getModelStructureByName(field.RelationalModel), mapIndex.(map[string]interface{}), modelField.Index(i), maestro, false)
				if err != nil {
					return nil, err
				}

				relationalArray = reflect.Append(relationalArray, reflect.ValueOf(relationalModel))
			}

			modelField.Set(relationalArray)
		} else if field.Type == FIELD_TYPE_RELATION_COL_IDS {
			fieldStructure := modelS.getFieldByJson(key)
			relationalModel := maestro.getModelStructureByName(fieldStructure.RelationalModel)
			modelFieldName := strings.Replace(field.Name, "IDs", "", -1)

			model1DB := reflect.New(reflect.TypeOf(modelS.Model)).Interface()

			if notFound := db.First(model1DB, modelID).RecordNotFound(); notFound {
				return nil, &[]JSendErrorDescription{{Code: "POSTED_ID_DONT_EXIST", Message: "This is an Erudito's error, please report"}}
			}

			for _, relIDInterface := range value.([]interface{}) {
				relID := int(relIDInterface.(float64))
				model2DB := reflect.New(reflect.TypeOf(relationalModel.Model)).Interface()

				if notFound := db.First(model2DB, relID).RecordNotFound(); notFound {
					return nil, &[]JSendErrorDescription{{Code: "RELATION_ID_DONT_EXISTS", Message: "Relation " + strconv.FormatInt(int64(relID), 10) + " of field " + fieldStructure.Name + " doesn't exists"}}
				}

				if err := db.Model(model1DB).Association(modelFieldName).Append(model2DB).Error; err != nil {
					return nil, &[]JSendErrorDescription{{Code: "RELATION_ID_DONT_EXISTS", Message: "Relation " + strconv.FormatInt(int64(relID), 10) + " of field " + fieldStructure.Name + " doesn't exists: " + err.Error()}}
				}
			}

			modelResponseElem.FieldByName(modelFieldName).Set(reflect.AppendSlice(modelResponseElem.FieldByName(modelFieldName), reflect.ValueOf(model1DB).Elem().FieldByName(modelFieldName)))
		}
	}

	return modelResponse.Interface(), nil
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
