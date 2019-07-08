package erudito

import (
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cristianob/erudito/nulls"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type (
	dbPollCallback func(r *http.Request, metaData MiddlewareMetaData) *gorm.DB

	maestro struct {
		router             *mux.Router
		models             map[string]modelStructure
		MiddlewaresInitial []MiddlewareInitial
		dBPoolCallback     dbPollCallback
	}

	RouteConfig struct {
		AcceptGET        bool
		AcceptPOST       bool
		AcceptPUT        bool
		AcceptPATCH      bool
		AcceptDELETE     bool
		AcceptCollection bool
	}
)

func CreateMaestro(router *mux.Router, dBPoolCallback dbPollCallback) *maestro {
	_maestro := new(maestro)
	_maestro.router = router.StrictSlash(true)
	_maestro.dBPoolCallback = dBPoolCallback
	_maestro.models = map[string]modelStructure{}
	_maestro.MiddlewaresInitial = []MiddlewareInitial{}

	return _maestro
}

func (m *maestro) GetRouter() *mux.Router {
	return m.router
}

func (m *maestro) getModelStructure(model Model) modelStructure {
	return m.models[reflect.TypeOf(model).Name()]
}

func (m *maestro) getModelStructureByName(name string) modelStructure {
	return m.models[name]
}

func (m *maestro) AddMiddlewareInitial(function MiddlewareInitialFunction, mwType int) {
	m.MiddlewaresInitial = append(m.MiddlewaresInitial, MiddlewareInitial{
		Function: function,
		Type:     mwType,
	})
}

func (m *maestro) AddModel(model Model, routeConfig RouteConfig) {
	m.checkModelIntegrity(model)

	modelStruct := m.generateModelStructure(model)
	m.models[modelStruct.Name] = modelStruct

	individualMethods := []string{}
	crudOptions := model.CRUDOptions()

	if routeConfig.AcceptGET {
		m.addGET("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" GET", model)
		individualMethods = append(individualMethods, "GET")
	}

	if routeConfig.AcceptPOST {
		m.addPOST("/"+crudOptions.ModelSingular, crudOptions.ModelSingular+" POST", model)
		m.AddOPTION("/"+crudOptions.ModelSingular, crudOptions.ModelSingular+" OPTION", []string{"POST"})
		m.checkRelations(model, crudOptions)
	}

	if routeConfig.AcceptPUT {
		m.addPUT("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" PUT", model)
		individualMethods = append(individualMethods, "PUT")
	}

	if routeConfig.AcceptPATCH {
		m.addPATCH("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" PATCH", model)
		individualMethods = append(individualMethods, "PATCH")
	}

	if routeConfig.AcceptDELETE {
		m.addDelete("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" DELETE", model)
		individualMethods = append(individualMethods, "DELETE")
	}

	if routeConfig.AcceptCollection {
		m.addCollection("/"+crudOptions.ModelPlural, crudOptions.ModelSingular+" Collection", model)
		m.AddOPTION("/"+crudOptions.ModelPlural, crudOptions.ModelSingular+" Collection", []string{"GET"})
		m.addCollectionCount("/"+crudOptions.ModelPlural+"/count", crudOptions.ModelSingular+" Collection Count", model)
		m.AddOPTION("/"+crudOptions.ModelPlural+"/count", crudOptions.ModelSingular+" Collection Count", []string{"GET"})
	}

	if len(individualMethods) > 0 {
		m.AddOPTION("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" Model OPTIONS", individualMethods)
	}
}

func (m *maestro) AddHealthCheck() {
	m.addRoute("GET", "/health", "HealthCheck", HealthCheckHandler(m))
	m.AddOPTION("/health", "HealthCheck OPTIONS", []string{"GET"})
	log.Println("[ERUDITO] Added route GET /health")
}

func (m *maestro) AddRoute(path, method, name string, handler func(dbPollCallback) http.HandlerFunc) {
	m.addRoute(method, path, name, handler(m.dBPoolCallback))
	log.Println("[ERUDITO] Added route " + method + " " + path)
}

func (m *maestro) addRoute(method, path, name string, handler http.HandlerFunc) {
	m.router.
		Methods(method).
		Path(path).
		Name(name).
		Handler(handler)
}

func (m *maestro) addCollection(path, name string, model Model) {
	m.addRoute("GET", path, name, CollectionHandler(model, m))
	log.Println("[ERUDITO] Added route GET " + path)
}

func (m *maestro) addCollectionCount(path, name string, model Model) {
	m.addRoute("GET", path, name, CollectionCountHandler(model, m))
	log.Println("[ERUDITO] Added route GET " + path)
}

func (m *maestro) addGET(path, name string, model Model) {
	m.addRoute("GET", path, name, GetHandler(model, m))
	log.Println("[ERUDITO] Added route GET " + path)
}

func (m *maestro) addPOST(path, name string, model Model) {
	m.addRoute("POST", path, name, PostHandler(model, m))
	log.Println("[ERUDITO] Added route POST " + path)
}

func (m *maestro) addRelation(path, name, fieldName string, model1, model2 Model) {
	m.addRoute("PUT", path, name, RelationAddHandler(model1, model2, fieldName, m))
	log.Println("[ERUDITO] Added route PUT " + path)
}

func (m *maestro) removeRelation(path, name, fieldName string, model1, model2 Model) {
	m.addRoute("DELETE", path, name, RelationRemoveHandler(model1, model2, fieldName, m))
	log.Println("[ERUDITO] Added route DELETE " + path)
}

func (m *maestro) addPUT(path, name string, model Model) {
	m.addRoute("PUT", path, name, PutHandler(model, m))
	log.Println("[ERUDITO] Added route PUT " + path)
}

func (m *maestro) addPATCH(path, name string, model Model) {
	m.addRoute("PATCH", path, name, PatchHandler(model, m))
	log.Println("[ERUDITO] Added route PATCH " + path)
}

func (m *maestro) addDelete(path, name string, model Model) {
	m.addRoute("DELETE", path, name, DeleteHandler(model, m))
	log.Println("[ERUDITO] Added route DELETE " + path)
}

func (m *maestro) AddOPTION(path, name string, methods []string) {
	m.addRoute("OPTIONS", path, name, OptionsHandler(methods))
}

func (m *maestro) checkModelIntegrity(model Model) {
	if reflect.TypeOf(model).Name() == "" || reflect.TypeOf(model).Kind() != reflect.Struct {
		log.Panic("Added Model needs to be a direct instance of a struct, not a reference")
	}
}

func (m *maestro) checkRelations(model Model, crudOptions CRUDOptions) {
	modelType := reflect.TypeOf(model)
	for i := 0; i < modelType.NumField(); i++ {
		if modelType.Field(i).Type.Kind() == reflect.Slice {
			gormTagset := modelType.Field(i).Tag.Get("gorm")
			if gormTagset == "" {
				continue
			}

			gormTags := strings.Split(gormTagset, ";")
			for _, tag := range gormTags {
				if !strings.HasPrefix(tag, "many2many") {
					continue
				}

				model2Type := modelType.Field(i).Type.Elem()
				if model2Type.Kind() == reflect.Ptr {
					model2Type = model2Type.Elem()
				}

				model2 := reflect.Zero(model2Type).Interface().(Model)
				crudOptions2 := model2.CRUDOptions()

				m.addRelation("/"+crudOptions.ModelSingular+"/{id1}/"+crudOptions2.ModelSingular+"/{id2}",
					crudOptions.ModelSingular+" - "+crudOptions2.ModelSingular+" Relation Add",
					modelType.Field(i).Name,
					model,
					model2,
				)

				m.removeRelation("/"+crudOptions.ModelSingular+"/{id1}/"+crudOptions2.ModelSingular+"/{id2}",
					crudOptions.ModelSingular+" - "+crudOptions2.ModelSingular+" Relation Remove",
					modelType.Field(i).Name,
					model,
					model2,
				)

				m.AddOPTION("/"+crudOptions.ModelSingular+"/{id1}/"+crudOptions2.ModelSingular+"/{id2}",
					crudOptions.ModelSingular+" - "+crudOptions2.ModelSingular+" Relation OPTIONS",
					[]string{"PUT", "DELETE"},
				)
			}
		}
	}
}

func (m *maestro) generateModelStructure(model Model) modelStructure {
	rtr := modelStructure{}

	modelType := reflect.TypeOf(model)
	crudOptions := model.CRUDOptions()

	rtr.Name = modelType.Name()
	rtr.Singular = crudOptions.ModelSingular
	rtr.Plural = crudOptions.ModelPlural
	rtr.Fields = []fieldStructure{}
	rtr.Model = model

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		if field.Anonymous {
			if field.Type == reflect.TypeOf(FullModel{}) {
				rtr.Type = MODEL_TYPE_FULL
			} else if field.Type == reflect.TypeOf(HardDeleteModel{}) {
				rtr.Type = MODEL_TYPE_HARDDELETE
			} else if field.Type == reflect.TypeOf(SimpleModel{}) {
				rtr.Type = MODEL_TYPE_SIMPLE
			}

			continue
		}

		rtrFieldStructure := fieldStructure{}
		rtrFieldStructure.Name = field.Name

		fieldJsonName := strings.Replace(field.Tag.Get("json"), ",omitempty", "", -1)
		if fieldJsonName == "" {
			continue
		}
		rtrFieldStructure.JsonName = fieldJsonName

		rtrType, rtrNullable, rtrRelationalModel, rtrError := m.getFieldType(field)
		if !rtrError {
			rtrFieldStructure.Type = rtrType
			rtrFieldStructure.Nullable = rtrNullable
			rtrFieldStructure.RelationalModel = rtrRelationalModel
		}

		if rtrType == FIELD_TYPE_RELATION_COL_MODELS {
			rtr.Fields = append(rtr.Fields, fieldStructure{
				Name:            rtrFieldStructure.Name + "IDs",
				JsonName:        rtrFieldStructure.JsonName + "_ids",
				Type:            FIELD_TYPE_RELATION_COL_IDS,
				Nullable:        rtrFieldStructure.Nullable,
				RelationalModel: rtrFieldStructure.RelationalModel,
				BaseField:       rtrFieldStructure.Name,
			})

			rtr.Fields = append(rtr.Fields, fieldStructure{
				Name:            rtrFieldStructure.Name + "IDsUpdate",
				JsonName:        rtrFieldStructure.JsonName + "_ids_update",
				Type:            FIELD_TYPE_RELATION_COL_IDS_UPDATE,
				Nullable:        rtrFieldStructure.Nullable,
				RelationalModel: rtrFieldStructure.RelationalModel,
				BaseField:       rtrFieldStructure.Name,
			})

			rtr.Fields = append(rtr.Fields, fieldStructure{
				Name:            rtrFieldStructure.Name + "IDsReplace",
				JsonName:        rtrFieldStructure.JsonName + "_ids_replace",
				Type:            FIELD_TYPE_RELATION_COL_IDS_REPLACE,
				Nullable:        rtrFieldStructure.Nullable,
				RelationalModel: rtrFieldStructure.RelationalModel,
				BaseField:       rtrFieldStructure.Name,
			})

			rtr.Fields = append(rtr.Fields, fieldStructure{
				Name:            rtrFieldStructure.Name + "Update",
				JsonName:        rtrFieldStructure.JsonName + "_update",
				Type:            FIELD_TYPE_RELATION_COL_MODELS_UPDATE,
				Nullable:        rtrFieldStructure.Nullable,
				RelationalModel: rtrFieldStructure.RelationalModel,
				BaseField:       rtrFieldStructure.Name,
			})

			rtr.Fields = append(rtr.Fields, fieldStructure{
				Name:            rtrFieldStructure.Name + "Replace",
				JsonName:        rtrFieldStructure.JsonName + "_replace",
				Type:            FIELD_TYPE_RELATION_COL_MODELS_REPLACE,
				Nullable:        rtrFieldStructure.Nullable,
				RelationalModel: rtrFieldStructure.RelationalModel,
				BaseField:       rtrFieldStructure.Name,
			})
		}

		rtr.Fields = append(rtr.Fields, rtrFieldStructure)
	}

	return rtr
}

func (m *maestro) getFieldType(field reflect.StructField) (int, bool, string, bool) {
	rtrType := 0
	rtrNullable := false
	rtrRelational := ""

	fieldType := field.Type
	eruditoType := field.Tag.Get("eruditoType")

	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	if fieldType == reflect.TypeOf(time.Time{}) {
		rtrType = FIELD_TYPE_COMMON_TIME
		rtrNullable = false
		return rtrType, rtrNullable, rtrRelational, false
	}

	nullableType := reflect.TypeOf((*nulls.Nullable)(nil)).Elem()
	if fieldType.Implements(nullableType) {
		if eruditoType == "JSON" {
			rtrType = FIELD_TYPE_COMMON_JSON
			rtrNullable = true
			return rtrType, rtrNullable, rtrRelational, false
		} else {
			rtrType = FIELD_TYPE_COMMON
			rtrNullable = true
			return rtrType, rtrNullable, rtrRelational, false
		}
	}

	if eruditoType == "JSON" {
		rtrType = FIELD_TYPE_COMMON_JSON
		return rtrType, rtrNullable, rtrRelational, false
	}

	if eruditoType == "REL_ID" {
		rtrType = FIELD_TYPE_RELATION_ID
		return rtrType, rtrNullable, rtrRelational, false
	}

	if fieldType.Kind() == reflect.Struct {
		if fieldType.Implements(reflect.TypeOf((*Model)(nil)).Elem()) {
			rtrType = FIELD_TYPE_RELATION_MODEL
			rtrRelational = fieldType.Name()
			return rtrType, rtrNullable, rtrRelational, false
		}

		return rtrType, rtrNullable, rtrRelational, true
	}

	if fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()

		if fieldType.Kind() == reflect.Ptr {
			rtrNullable = true
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() != reflect.Struct {
			return rtrType, rtrNullable, rtrRelational, true
		}

		if fieldType.Implements(reflect.TypeOf((*Model)(nil)).Elem()) {
			rtrType = FIELD_TYPE_RELATION_COL_MODELS
			rtrRelational = fieldType.Name()
			return rtrType, rtrNullable, rtrRelational, false
		}

		return rtrType, rtrNullable, rtrRelational, true
	}

	rtrType = FIELD_TYPE_COMMON
	return rtrType, rtrNullable, rtrRelational, false
}
