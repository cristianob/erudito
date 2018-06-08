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
	Code    string `json:"code"`
	Message string `json:"message"`
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

func MakeSingularDataStruct(dataType reflect.Type, data interface{}) interface{} {
	sfs := []reflect.StructField{
		{
			Name: dataType.Name(),
			Type: reflect.ValueOf(data).Type(),
			Tag:  reflect.StructTag("json:\"" + reflect.ValueOf(data).MethodByName("ModelSingular").Call([]reflect.Value{})[0].String() + "\""),
		},
	}

	sr := reflect.ValueOf(reflect.New(reflect.StructOf(sfs)).Interface())
	sr.Elem().Field(0).Set(reflect.ValueOf(data))

	return sr.Interface()
}

func MakeArrayDataStruct(dataType reflect.Type, data interface{}) interface{} {
	sfs := []reflect.StructField{
		{
			Name: dataType.Name(),
			Type: reflect.SliceOf(dataType),
			Tag:  reflect.StructTag("json:\"" + reflect.Zero(dataType).MethodByName("ModelPlural").Call([]reflect.Value{})[0].String() + "\""),
		},
	}

	sr := reflect.ValueOf(reflect.New(reflect.StructOf(sfs)).Interface())
	sr.Elem().Field(0).Set(reflect.ValueOf(data).Elem())

	return sr.Interface()
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
