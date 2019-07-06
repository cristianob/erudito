package erudito

import (
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
)

/*
 * Type
 */
const (
	MIDDLEWARE_TYPE_GLOBAL     = 1
	MIDDLEWARE_TYPE_COLLECTION = 2
	MIDDLEWARE_TYPE_GET        = 3
	MIDDLEWARE_TYPE_POST       = 4
	MIDDLEWARE_TYPE_PUT        = 5
	MIDDLEWARE_TYPE_PATCH      = 6
	MIDDLEWARE_TYPE_DELETE     = 7
)

/*
 * Level
 */
const (
	MIDDLEWARE_LEVEL_ANY  = 1
	MIDDLEWARE_LEVEL_ROOT = 2
)

type (
	MidlewareMetaData map[string]interface{}

	MiddlewarePREFunction func(data MiddlewarePREData) MiddlewarePREReturn

	MiddlewarePRE struct {
		Function MiddlewarePREFunction
		Type     int
		Level    int
	}

	MiddlewarePREData struct {
		R      *http.Request
		W      http.ResponseWriter
		DbConn *gorm.DB
		Meta   MidlewareMetaData
		Model  reflect.Value
	}

	MiddlewarePREReturn struct {
		R     *http.Request
		W     http.ResponseWriter
		Meta  MidlewareMetaData
		Model reflect.Value
		Error *[]JSendErrorDescription
	}
)

func MiddlewarePRECreate(function MiddlewarePREFunction, Type int, Level int) MiddlewarePRE {
	return MiddlewarePRE{
		Function: function,
		Type:     Type,
		Level:    Level,
	}
}
