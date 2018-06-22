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
	router         *mux.Router
	models         []interface{}
	dBPoolCallback func(r *http.Request) *gorm.DB
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

func (m *maestro) AddModel(model Model) {
	if reflect.TypeOf(model).Name() == "" || reflect.TypeOf(model).Kind() != reflect.Struct {
		log.Panic("Added Model needs to be a direct instance of a struct, not a reference")
	}

	individualMethods := []string{}

	if model.AcceptGET() {
		m.router.
			Methods("GET").
			Path("/" + model.ModelSingular() + "/{id}").
			Name(model.ModelSingular() + " GET").
			Handler(GetHandler(model, m.dBPoolCallback))

		individualMethods = append(individualMethods, "GET")

		log.Println("Added route GET /" + model.ModelSingular() + "/{id}")
	}

	if model.AcceptPOST() {
		m.router.
			Methods("POST").
			Path("/" + model.ModelSingular()).
			Name(model.ModelSingular() + " POST").
			Handler(PostHandler(model, m.dBPoolCallback))

		m.router.
			Methods("OPTIONS").
			Path("/" + model.ModelSingular()).
			Name(model.ModelSingular() + " OPTION").
			Handler(OptionsHandler([]string{"POST"}))

		log.Println("Added route POST /" + model.ModelSingular())

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

					model2 := reflect.Zero(modelType.Field(i).Type.Elem()).Interface().(Model)

					m.router.
						Methods("PUT").
						Path("/" + model.ModelSingular() + "/{id1}/" + model2.ModelSingular() + "/{id2}").
						Name(model.ModelSingular() + " - " + model2.ModelSingular() + " Relation Add").
						Handler(RelationAddHandler(model, model2, modelType.Field(i).Name, m.dBPoolCallback))

					log.Println("Added route PUT /" + model.ModelSingular() + "/{id1}/" + model2.ModelSingular() + "/{id2}")

					m.router.
						Methods("DELETE").
						Path("/" + model.ModelSingular() + "/{id1}/" + model2.ModelSingular() + "/{id2}").
						Name(model.ModelSingular() + " - " + model2.ModelSingular() + " Relation Remove").
						Handler(RelationRemoveHandler(model, model2, modelType.Field(i).Name, m.dBPoolCallback))

					log.Println("Added route DELETE /" + model.ModelSingular() + "/{id1}/" + model2.ModelSingular() + "/{id2}")

					m.router.
						Methods("OPTIONS").
						Path("/" + model.ModelSingular() + "/{id1}/" + model2.ModelSingular() + "/{id2}").
						Name(model.ModelSingular() + " - " + model2.ModelSingular() + " Relation OPTIONS").
						Handler(OptionsHandler([]string{"PUT", "DELETE"}))
				}
			}
		}
	}

	if model.AcceptPUT() {
		m.router.
			Methods("PUT").
			Path("/" + model.ModelSingular() + "/{id}").
			Name(model.ModelSingular() + " PUT").
			Handler(PutHandler(model, m.dBPoolCallback))

		individualMethods = append(individualMethods, "PUT")

		log.Println("Added route PUT /" + model.ModelSingular() + "/{id}")
	}

	modelType := reflect.TypeOf(model)
	for i := 0; i < modelType.NumField(); i++ {
		if modelType.Field(i).Type.Kind() != reflect.Slice &&
			modelType.Field(i).Type.Kind() != reflect.Struct &&
			checkIfTagExists("microUpdate", modelType.Field(i).Tag.Get("erudito")) {

			m.router.
				Methods("PUT").
				Path("/" + model.ModelSingular() + "/{id}/" + modelType.Field(i).Tag.Get("json")).
				Name(model.ModelSingular() + " - " + modelType.Field(i).Name + " Micro Update").
				Handler(MicroUpdateHandler(model, modelType.Field(i).Tag.Get("json"), m.dBPoolCallback))

			m.router.
				Methods("OPTIONS").
				Path("/" + model.ModelSingular() + "/{id}/" + modelType.Field(i).Tag.Get("json")).
				Name(model.ModelSingular() + " - " + modelType.Field(i).Name + " Micro Update OPTIONS").
				Handler(OptionsHandler([]string{"PUT"}))

			log.Println("Added route PUT /" + model.ModelSingular() + "/{id}/" + modelType.Field(i).Tag.Get("json"))

		}
	}

	if model.AcceptDELETE() {
		m.router.
			Methods("DELETE").
			Path("/" + model.ModelSingular() + "/{id}").
			Name(model.ModelSingular() + " DELETE").
			Handler(DeleteHandler(model, m.dBPoolCallback))

		individualMethods = append(individualMethods, "DELETE")

		log.Println("Added route DELETE /" + model.ModelSingular() + "/{id}")
	}

	if model.AcceptCollection() {
		m.router.
			Methods("GET").
			Path("/" + model.ModelPlural()).
			Name(model.ModelSingular() + " Collection").
			Handler(CollectionHandler(model, m.dBPoolCallback))

		m.router.
			Methods("OPTIONS").
			Path("/" + model.ModelPlural()).
			Name(model.ModelSingular() + " Collection").
			Handler(OptionsHandler([]string{"GET"}))

		log.Println("Added route GET /" + model.ModelPlural())

		m.router.
			Methods("GET").
			Path("/" + model.ModelPlural() + "/count").
			Name(model.ModelSingular() + " Collection Count").
			Handler(CollectionCountHandler(model, m.dBPoolCallback))

		m.router.
			Methods("OPTIONS").
			Path("/" + model.ModelPlural() + "/count").
			Name(model.ModelSingular() + " Collection Count").
			Handler(OptionsHandler([]string{"GET"}))

		log.Println("Added route GET /" + model.ModelPlural() + "/count")
	}

	if len(individualMethods) > 0 {
		m.router.
			Methods("OPTIONS").
			Path("/" + model.ModelSingular() + "/{id}").
			Name(model.ModelSingular() + " Model OPTIONS").
			Handler(OptionsHandler(individualMethods))
	}
}

func (m *maestro) AddHealthCheck() {
	m.router.
		Methods("GET").
		Path("/health").
		Name("HealthCheck").
		Handler(HealthCheckHandler(m.dBPoolCallback))

	m.router.
		Methods("OPTIONS").
		Path("/health").
		Name("HealthCheck OPTIONS").
		Handler(OptionsHandler([]string{"GET"}))

	log.Println("Added route GET /health")
}

func (m *maestro) AddRoute(path, method, name string, handler func(func(r *http.Request) *gorm.DB) http.Handler) {
	m.router.
		Methods(method).
		Path(path).
		Name(name).
		Handler(handler(m.dBPoolCallback))

	m.router.
		Methods("OPTIONS").
		Path(path).
		Name(name + " OPTIONS").
		Handler(OptionsHandler([]string{method}))

	log.Println("Added route " + method + " " + path)
}
