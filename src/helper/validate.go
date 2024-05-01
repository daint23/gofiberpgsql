package helper

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Param string `json:"value,omitempty"`
}

func ValidateStruct[T any](payload T, validate *validator.Validate) error {
	var errFields []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = strings.ToLower(err.Field())
			element.Tag = err.Tag()
			element.Param = err.Param()
			errFields = append(errFields, &element)
		}
	}
	if len(errFields) == 0 {
		return nil
	}
	marshaledErr, _ := json.Marshal(&errFields)
	return errors.New(string(marshaledErr))
}
