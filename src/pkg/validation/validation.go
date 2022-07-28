package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func Errors(err error) interface{} {
	var m map[string]interface{} = make(map[string]interface{})
	m["status"] = false
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		m["message"] = "invalid request body"
		return m
	}
	m["message"] = "validation failed"
	m["error"] = make(map[string]string)
	for _, err := range validationErrors {
		m["error"].(map[string]string)[strings.ToLower(err.Field())] = getErrorMessage(err)
	}
	return m
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "email":
		return "email is invalid"
	case "required":
		return "field is required"
	case "min":
		switch err.Field() {
		case "Phone":
			return "minimium lenght is 8"
		case "Password":
			return "minimium lenght is 6"
		default:
			return "lenght is too short"
		}
	case "number":
		return "numeric required"
	default:
		return fmt.Sprintf("validation failed for %s", err.Tag())
	}
}
