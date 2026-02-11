package pkg

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) []string {
	var errors []string

	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var message string
			switch err.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", err.Field())
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
			case "max":
				message = fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param())
			case "gt":
				message = fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
			default:
				message = fmt.Sprintf("%s is invalid", err.Field())
			}
			errors = append(errors, message)
		}
	}

	return errors
}
