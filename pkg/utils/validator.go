package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(data interface{}) map[string]string {
	err := validate.Struct(data)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErrors {
			errors[err.Field()] = getSimpleErrorMessage(err)
		}
	}

	return errors
}

func getSimpleErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Minimum length is %s", err.Param())
	case "max":
		return fmt.Sprintf("Maximum length is %s", err.Param())
	case "len":
		return fmt.Sprintf("Must be exactly %s characters", err.Param())
	case "oneof":
		options := strings.ReplaceAll(err.Param(), " ", ", ")
		return fmt.Sprintf("Must be one of: %s", options)
	case "uuid":
		return "Must be a valid UUID"
	default:
		return fmt.Sprintf("Invalid %s field", err.Field())
	}
}

func FormatValidationErrors(errors map[string]string) string {
	var msgs []string
	for field, msg := range errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(msgs, "; ")
}
