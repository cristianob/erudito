package erudito

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
)

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

func SendEmptyResponse(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

func SendData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := new(JSendSuccess)
	response.Status = "success"
	response.Data = data

	json.NewEncoder(w).Encode(response)
}

func SendError(w http.ResponseWriter, status int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := new(JSendError)
	response.Status = "error"
	response.Data = []JSendErrorDescription{
		{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(response)
}
