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

	if model.AcceptGET() {
		m.router.
			Methods("GET").
			Path("/" + model.ModelSingular() + "/{id}").
			Name(model.ModelSingular() + " GET").
			Handler(GetHandler(model, m.dBPoolCallback))

		log.Println("Added route GET /" + model.ModelSingular() + "/{id}")
	}

	if model.AcceptPOST() {
		m.router.
			Methods("POST").
			Path("/" + model.ModelSingular()).
			Name(model.ModelSingular() + " POST").
			Handler(PostHandler(model, m.dBPoolCallback))

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

		log.Println("Added route PUT /" + model.ModelSingular() + "/{id}")
	}

	if model.AcceptDELETE() {
		m.router.
			Methods("DELETE").
			Path("/" + model.ModelSingular() + "/{id}").
			Name(model.ModelSingular() + " DELETE").
			Handler(DeleteHandler(model, m.dBPoolCallback))

		log.Println("Added route DELETE /" + model.ModelSingular() + "/{id}")
	}

	if model.AcceptCollection() {
		m.router.
			Methods("GET").
			Path("/" + model.ModelPlural()).
			Name(model.ModelSingular() + " Collection").
			Handler(CollectionHandler(model, m.dBPoolCallback))

		log.Println("Added route GET /" + model.ModelPlural())

		m.router.
			Methods("GET").
			Path("/" + model.ModelPlural() + "/count").
			Name(model.ModelSingular() + " Collection Count").
			Handler(CollectionCountHandler(model, m.dBPoolCallback))

		log.Println("Added route GET /" + model.ModelPlural() + "/count")
	}
}
