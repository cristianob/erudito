package erudito

import (
	"net/http"
)

func HealthCheckHandler(maestro *maestro) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		SendData(w, http.StatusOK, "OK")
	})
}
