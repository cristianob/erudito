package erudito

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"

	"github.com/gorilla/mux"
)

type JSendSuccess struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type JSendError struct {
	Status string                  `json:"status"`
	Data   []JSendErrorDescription `json:"data"`
}

type JSendErrorDescription struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Refer   *string `json:"refer,omitempty"`
	Pos     *uint   `json:"pos,omitempty"`
}

type JSendFail struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func GetNumericRouteField(r *http.Request, field string) (uint, error) {
	fieldV := mux.Vars(r)[field]
	fieldVNumeric, err := strconv.Atoi(fieldV)

	if err != nil {
		return 0, err
	}

	return uint(fieldVNumeric), nil
}

func MakeSingularDataStruct(dataType reflect.Type, data interface{}, modelS modelStructure) interface{} {
	valueOfData := reflect.ValueOf(data)

	sfs := []reflect.StructField{
		{
			Name: dataType.Name(),
			Type: valueOfData.Type(),
			Tag:  reflect.StructTag("json:\"" + modelS.Singular + "\""),
		},
	}

	sr := reflect.ValueOf(reflect.New(reflect.StructOf(sfs)).Interface())
	sr.Elem().Field(0).Set(valueOfData)

	return sr.Interface()
}

func MakeArrayDataStruct(dataType reflect.Type, data []interface{}, modelS modelStructure) interface{} {
	sfs := []reflect.StructField{
		{
			Name: dataType.Name(),
			Type: reflect.TypeOf([]interface{}{}),
			Tag:  reflect.StructTag("json:\"" + modelS.Plural + "\""),
		},
		{
			Name: "Count",
			Type: reflect.TypeOf(0),
			Tag:  reflect.StructTag("json:\"count\""),
		},
	}

	sr := reflect.ValueOf(reflect.New(reflect.StructOf(sfs)).Interface())

	sr.Elem().Field(0).Set(reflect.ValueOf(data))
	sr.Elem().Field(1).Set(reflect.ValueOf(reflect.ValueOf(data).Len()))

	return sr.Interface()
}

func AddCORSHeaders(w http.ResponseWriter, methods string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", methods)
	w.Header().Add("Access-Control-Allow-Headers", "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization")
	w.Header().Add("Access-Control-Max-Age", "1728000")
}

func SendEmptyResponse(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func SendData(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := new(JSendSuccess)
	response.Status = "success"
	response.Data = data

	json.NewEncoder(w).Encode(response)
}

func SendError(w http.ResponseWriter, status int, errs []JSendErrorDescription) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := new(JSendError)
	response.Status = "error"
	response.Data = errs

	json.NewEncoder(w).Encode(response)
}

func SendSingleError(w http.ResponseWriter, status int, message string, code string) {
	SendError(w, status, []JSendErrorDescription{
		{
			Code:    code,
			Message: message,
		},
	})
}

func SendFail(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := new(JSendFail)
	response.Status = "fail"
	response.Message = message

	json.NewEncoder(w).Encode(response)
}

func InternalError(w http.ResponseWriter, err error) {
	SendFail(w, http.StatusInternalServerError, "An internal error ocourred, please contact the system administrator")
	debug.PrintStack()
	log.Println("Error: ", err.Error())
}
