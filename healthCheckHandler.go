package erudito

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

func HealthCheckHandler(DBPoolCallback func(r *http.Request) *gorm.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddCORSHeaders(w, "GET")

		db := DBPoolCallback(r)
		if db == nil {
			SendSingleError(w, http.StatusInternalServerError, "Database error!", "DATABASE_ERROR")
			return
		}

		SendEmptyResponse(w, http.StatusOK)
	})
}
