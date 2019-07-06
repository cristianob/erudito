package main

import (
	"log"

	"github.com/cristianob/erudito"
)

var globalMiddleware1 erudito.MiddlewarePRE
var blockMiddleware erudito.MiddlewarePRE
var middlewareA1 erudito.MiddlewarePRE
var middlewareB1 erudito.MiddlewarePRE

func createMiddlewares() {
	globalMiddleware1 = erudito.MiddlewarePRECreate(func(data erudito.MiddlewarePREData) erudito.MiddlewarePREReturn {
		log.Println("This middleware just fowards")

		return erudito.MiddlewarePREReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: nil,
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)

	blockMiddleware = erudito.MiddlewarePRECreate(func(data erudito.MiddlewarePREData) erudito.MiddlewarePREReturn {
		log.Println("This middleware blocks")

		return erudito.MiddlewarePREReturn{
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

	middlewareA1 = erudito.MiddlewarePRECreate(func(data erudito.MiddlewarePREData) erudito.MiddlewarePREReturn {
		log.Println("This middleware changes model A")

		data.Model.FieldByName("Field1").SetString("test_modified")

		return erudito.MiddlewarePREReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: nil,
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)

	middlewareB1 = erudito.MiddlewarePRECreate(func(data erudito.MiddlewarePREData) erudito.MiddlewarePREReturn {
		log.Println("This middleware changes model B")

		data.Model.FieldByName("Field1").SetString("test_modified_b")

		return erudito.MiddlewarePREReturn{
			R:     data.R,
			W:     data.W,
			Meta:  data.Meta,
			Model: data.Model,
			Error: nil,
		}
	}, erudito.MIDDLEWARE_TYPE_GLOBAL, erudito.MIDDLEWARE_LEVEL_ANY)
}
