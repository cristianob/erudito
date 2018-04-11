package erudito

import (
	"github.com/gorilla/mux"
)

func CreateRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	// for _, route := range routes {
	// 	var innerHandler HttpDBHandler
	// 	innerHandler = route.HandlerFunc

	// 	var handler http.Handler
	// 	handler = RouteDecorator(innerHandler, route.Name)

	// 	router.
	// 		Methods(route.Method).
	// 		Path(route.Pattern).
	// 		Name(route.Name).
	// 		Handler(handler)

	// }
	return router
}
