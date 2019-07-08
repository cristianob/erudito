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
	MIDDLEWARE_TYPE_RELATION   = 8
)

/*
 * Level
 */
const (
	MIDDLEWARE_LEVEL_ANY  = 1
	MIDDLEWARE_LEVEL_ROOT = 2
)

type MiddlewareMetaData map[string]interface{}

/*
 * MIDDLEWARE Initial
 */

type (
	MiddlewareInitialFunction func(data MiddlewareInitialData) MiddlewareInitialReturn

	MiddlewareInitial struct {
		Function MiddlewareInitialFunction
		Type     int
	}

	MiddlewareInitialData struct {
		R    *http.Request
		W    http.ResponseWriter
		Meta MiddlewareMetaData
	}

	MiddlewareInitialReturn struct {
		R     *http.Request
		W     http.ResponseWriter
		Meta  MiddlewareMetaData
		Error *[]JSendErrorDescription
	}
)

func MiddlewareInitialCreate(function MiddlewareInitialFunction, Type int, Level int) MiddlewareInitial {
	return MiddlewareInitial{
		Function: function,
		Type:     Type,
	}
}

/*
 * MIDDLEWARE Before
 */

type (
	MiddlewareBeforeFunction func(data MiddlewareBeforeData) MiddlewareBeforeReturn

	MiddlewareBefore struct {
		Function MiddlewareBeforeFunction
		Type     int
		Level    int
	}

	MiddlewareBeforeData struct {
		R      *http.Request
		W      http.ResponseWriter
		DbConn *gorm.DB
		Meta   MiddlewareMetaData
		Model  reflect.Value
	}

	MiddlewareBeforeReturn struct {
		R     *http.Request
		W     http.ResponseWriter
		Meta  MiddlewareMetaData
		Model reflect.Value
		Error *[]JSendErrorDescription
	}
)

func MiddlewareBeforeCreate(function MiddlewareBeforeFunction, Type int, Level int) MiddlewareBefore {
	return MiddlewareBefore{
		Function: function,
		Type:     Type,
		Level:    Level,
	}
}

/*
 * MIDDLEWARE After
 */

type (
	MiddlewareAfterFunction func(data MiddlewareAfterData) MiddlewareAfterReturn

	MiddlewareAfter struct {
		Function MiddlewareAfterFunction
		Type     int
		Level    int
	}

	MiddlewareAfterData struct {
		R        *http.Request
		W        http.ResponseWriter
		DbConn   *gorm.DB
		Meta     MiddlewareMetaData
		Response map[string]interface{}
	}

	MiddlewareAfterReturn struct {
		R        *http.Request
		W        http.ResponseWriter
		Meta     MiddlewareMetaData
		Response map[string]interface{}
		Error    *[]JSendErrorDescription
	}
)

func MiddlewareAfterCreate(function MiddlewareAfterFunction, Type int, Level int) MiddlewareAfter {
	return MiddlewareAfter{
		Function: function,
		Type:     Type,
		Level:    Level,
	}
}
