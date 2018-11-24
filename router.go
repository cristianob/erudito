package erudito

import (
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type maestro struct {
	router                *mux.Router
	models                []interface{}
	beforeRequestCallback func(r *http.Request) []JSendErrorDescription
	dBPoolCallback        func(r *http.Request) *gorm.DB
}

func CreateMaestro(fn func(r *http.Request) *gorm.DB) *maestro {
	_maestro := new(maestro)
	_maestro.router = mux.NewRouter().StrictSlash(true)
	_maestro.dBPoolCallback = fn

	return _maestro
}

func (m *maestro) GetRouter() *mux.Router {
	return m.router
}

func (m *maestro) SetBeforeRequestCallback(cb func(r *http.Request) []JSendErrorDescription) {
	m.beforeRequestCallback = cb
}

func (m *maestro) AddModel(model Model) {
	if reflect.TypeOf(model).Name() == "" || reflect.TypeOf(model).Kind() != reflect.Struct {
		log.Panic("Added Model needs to be a direct instance of a struct, not a reference")
	}

	individualMethods := []string{}

	crudOptions := model.CRUDOptions()

	if crudOptions.AcceptGET {
		m.addGET("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" GET", model)
		individualMethods = append(individualMethods, "GET")
	}

	if crudOptions.AcceptPOST {
		m.addPOST("/"+crudOptions.ModelSingular, crudOptions.ModelSingular+" POST", model)
		m.addOPTION("/"+crudOptions.ModelSingular, crudOptions.ModelSingular+" OPTION", []string{"POST"})

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

					m.addOPTION("/"+crudOptions.ModelSingular+"/{id1}/"+crudOptions2.ModelSingular+"/{id2}",
						crudOptions.ModelSingular+" - "+crudOptions2.ModelSingular+" Relation OPTIONS",
						[]string{"PUT", "DELETE"},
					)
				}
			}
		}
	}

	if crudOptions.AcceptPUT {
		m.addPUT("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" PUT", model)
		individualMethods = append(individualMethods, "PUT")
	}

	if crudOptions.AcceptPATCH {
		m.addPATCH("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" PATCH", model)
		individualMethods = append(individualMethods, "PATCH")
	}

	if crudOptions.AcceptDELETE {
		m.addDelete("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" DELETE", model)
		individualMethods = append(individualMethods, "DELETE")
	}

	if crudOptions.AcceptCollection {
		m.addCollection("/"+crudOptions.ModelPlural, crudOptions.ModelSingular+" Collection", model)
		m.addOPTION("/"+crudOptions.ModelPlural, crudOptions.ModelSingular+" Collection", []string{"GET"})
		m.addCollectionCount("/"+crudOptions.ModelPlural+"/count", crudOptions.ModelSingular+" Collection Count", model)
		m.addOPTION("/"+crudOptions.ModelPlural+"/count", crudOptions.ModelSingular+" Collection Count", []string{"GET"})
	}

	if len(individualMethods) > 0 {
		m.addOPTION("/"+crudOptions.ModelSingular+"/{id}", crudOptions.ModelSingular+" Model OPTIONS", individualMethods)
	}
}

func (m *maestro) AddHealthCheck() {
	m.addRoute("GET", "/health", "HealthCheck", HealthCheckHandler(m))
	m.addOPTION("/health", "HealthCheck OPTIONS", []string{"GET"})
	log.Println("[ERUDITO] Added route GET /health")
}

func (m *maestro) AddRoute(path, method, name string, handler func(func(r *http.Request) *gorm.DB) http.HandlerFunc) {
	m.addRoute(method, path, name, handler(m.dBPoolCallback))
	m.addOPTION(path, name+" OPTIONS", []string{method})
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

func (m *maestro) addOPTION(path, name string, methods []string) {
	m.addRoute("OPTIONS", path, name, OptionsHandler(methods))
}
