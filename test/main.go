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
	maestro.AddModel(A{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(B{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(C{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(D{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddModel(E{}, erudito.RouteConfig{AcceptGET: true, AcceptPOST: true, AcceptPUT: true, AcceptPATCH: true, AcceptDELETE: true, AcceptCollection: true})
	maestro.AddHealthCheck()

	log.Println("Initializing Server in port 80")
	log.Fatal(http.ListenAndServe(":80", router))
}

// func beforeRequest(r *http.Request) ([]erudito.JSendErrorDescription, map[string]interface{}) {
// 	authKey := auth.GetAuthKey(r)
// 	user := auth.AuthCheck(authKey)

// 	if user == nil {
// 		return []erudito.JSendErrorDescription{
// 			{
// 				Code:    "NOT_AUTHORIZED",
// 				Message: "User not authorized",
// 			},
// 		}, nil
// 	}

// 	r.Header.Set("X-Auth-User", strconv.FormatUint(user.UserID, 10))
// 	r.Header.Set("X-Auth-Group", strconv.FormatUint(user.UserGroup, 10))

// 	return nil, map[string]interface{}{
// 		"User": user,
// 	}
// }
