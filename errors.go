package erudito

type FieldError struct {
	FieldName string `json:"field_name`
	Message   string `json:"message`
}

func (e FieldError) Error() string {
	return "Field `" + e.FieldName + "` is invalid: " + e.Message
}
