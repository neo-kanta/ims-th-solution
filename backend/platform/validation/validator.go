package validation

import (
	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator for request validation.
var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidateStruct validates a struct using go-playground/validator tags.
// Returns nil if valid, or a slice of field error descriptions.
func ValidateStruct(s interface{}) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   e.Field(),
			Tag:     e.Tag(),
			Value:   e.Param(),
			Message: formatMessage(e),
		})
	}
	return fieldErrors
}

// FieldError describes a single field validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "min":
		return e.Field() + " must be at least " + e.Param()
	case "max":
		return e.Field() + " must be at most " + e.Param()
	case "email":
		return e.Field() + " must be a valid email address"
	case "uuid":
		return e.Field() + " must be a valid UUID"
	default:
		return e.Field() + " failed validation: " + e.Tag()
	}
}
