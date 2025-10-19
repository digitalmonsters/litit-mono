package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Run("should return nil if the given struct is valid", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name" validate:"required"`
		}
		Init(&Option{})

		errors := Validate(TestStruct{Name: "test"})
		assert.Nil(t, errors)
	})

	t.Run("should return a map of errors if the given struct is invalid", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name" validate:"required"`
			Email string `json:"email" validate:"required,email"`
		}
		Init(&Option{})

		errors := Validate(TestStruct{})
		assert.Equal(t, errors, map[string]string{
			"name":  "name is required.",
			"email": "email is required.",
		})
	})
}
