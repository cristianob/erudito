package erudito

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
