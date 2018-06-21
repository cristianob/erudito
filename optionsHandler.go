package erudito

import (
	"log"
	"net/http"
	"strings"
)

func OptionsHandler(headers []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", strings.Join(headers, ", "))
		w.Header().Add("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization")
		w.Header().Add("Access-Control-Max-Age", "1728000")

		log.Println(w)

		SendEmptyResponse(w, http.StatusOK)
	})
}
