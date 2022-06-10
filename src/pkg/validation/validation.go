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
		m["error"].(map[string]string)[strings.ToLower(err.Field())] = getErrorMessage(err.Tag())
	}
	return m
}

func getErrorMessage(tag string) string {
	switch tag {
	case "email":
		return "email is invalid"
	default:
		return fmt.Sprintf("validation failed for %s", tag)
	}
}
