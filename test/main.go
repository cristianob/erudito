package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cristianob/erudito"
)

func main() {
	log.Println("===== ERUDITO TEST =====")

	databaseInit()
	databaseMigrate()
	defer databaseClose()

	createMiddlewares()

	router := mux.NewRouter()

	maestro := erudito.CreateMaestro(router, databaseResolve)
	maestro.AddMiddlewareInitial(SomeInitialMiddleware, erudito.MIDDLEWARE_TYPE_GLOBAL)
	maestro.AddModel(A{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(B{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(C{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(D{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(E{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddHealthCheck()

	log.Println("Initializing Server in port 80")
	log.Fatal(http.ListenAndServe(":80", router))
}

func SomeInitialMiddleware(data erudito.MiddlewareInitialData) erudito.MiddlewareInitialReturn {
	log.Println("Before everything!")

	metaData := data.Meta
	metaData["Initial"] = "Initial store data here"

	return erudito.MiddlewareInitialReturn{
		R:     data.R,
		W:     data.W,
		Meta:  metaData,
		Error: nil,
	}
}
