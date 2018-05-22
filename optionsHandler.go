package erudito

import (
	"net/http"
	"strings"
)

func OptionsHandler(headers []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Access-Control-Allow-Origin", "*")
		r.Header.Add("Access-Control-Allow-Credentials", "true")
		r.Header.Add("Access-Control-Allow-Methods", strings.Join(headers, ", "))
		r.Header.Add("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization")
		r.Header.Add("Access-Control-Max-Age", "1728000")

		SendEmptyResponse(w, http.StatusOK)
	})
}
