package main

import (
	"log"

	"github.com/cristianob/erudito"
)

var globalMiddleware1 erudito.MiddlewareBefore
var blockMiddleware erudito.MiddlewareBefore
var middlewareA1 erudito.MiddlewareBefore
var middlewareA2 erudito.MiddlewareAfter
var middlewareB1 erudito.MiddlewareBefore

func createMiddlewares() {
	globalMiddleware1 = erudito.MiddlewareBeforeCreate(func(data erudito.MiddlewareBeforeData) erudito.MiddlewareBeforeReturn {
		log.Println("This middleware fowards and prints the Initial Meta: " + data.Meta["Initial"].(string))

		return erudito.MiddlewareBeforeReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: nil,
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)

	blockMiddleware = erudito.MiddlewareBeforeCreate(func(data erudito.MiddlewareBeforeData) erudito.MiddlewareBeforeReturn {
		log.Println("This middleware blocks")

		return erudito.MiddlewareBeforeReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: &[]erudito.JSendErrorDescription{
				{
					Code:    "GANDALF",
					Message: "You shall not pass",
				},
			},
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)

	middlewareA1 = erudito.MiddlewareBeforeCreate(func(data erudito.MiddlewareBeforeData) erudito.MiddlewareBeforeReturn {
		log.Println("This middleware changes model A")

		data.Model.FieldByName("Field1").SetString("test_modified")

		return erudito.MiddlewareBeforeReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: nil,
		}
	}, erudito.MIDDLEWARE_TYPE_POST, erudito.MIDDLEWARE_LEVEL_ANY)

	middlewareA2 = erudito.MiddlewareAfterCreate(func(data erudito.MiddlewareAfterData) erudito.MiddlewareAfterReturn {
		log.Println("This middleware changes the return model A")

		data.Response["field1"] = "only_return_will_be_that"

		return erudito.MiddlewareAfterReturn{
			R:        data.R,
			W:        data.W,
			Meta:     data.Meta,
			Response: data.Response,
			Error:    nil,
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)

	middlewareB1 = erudito.MiddlewareBeforeCreate(func(data erudito.MiddlewareBeforeData) erudito.MiddlewareBeforeReturn {
		log.Println("This middleware changes model B")

		data.Model.FieldByName("Field1").SetString("test_modified_b")

		return erudito.MiddlewareBeforeReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: nil,
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)
}
