package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

const (
	defaultFieldNameAccessor = "json"
)

// FieldError is a field error
type FieldError struct {
	FieldName string
	validator.FieldError
}

// MessageFormatter is a message formatter
type MessageFormatter map[string]func(err *FieldError) string

// Validator is a struct validator
type Validator struct {
	fieldNameAccessor string
	messageFormatter  MessageFormatter
}

var _validator *Validator

var defaultFieldHandler = func(err *FieldError) string {
	return fmt.Sprintf("Invalid value for '%s'.", strings.ToLower(err.FieldName))
}

var defaultMessageFormatter = MessageFormatter{
	"required": func(err *FieldError) string {
		return fmt.Sprintf("%s is required.", err.FieldName)
	},
	"email": func(err *FieldError) string {
		return "Invalid email address."
	},
	"min": func(err *FieldError) string {
		return fmt.Sprintf("%s must be at least %s characters long.", err.FieldName, err.Param())
	},
	"max": func(err *FieldError) string {
		return fmt.Sprintf("%s must be at most %s characters long.", err.FieldName, err.Param())
	},
	"oneof": func(err *FieldError) string {
		return fmt.Sprintf("%s must be one of %s.", err.FieldName, err.Param())
	},
}

type Option struct {
	FieldNameAccessor string
	MessageFormatter
}

// Init initializes the validator
func Init(option *Option) {
	if option.FieldNameAccessor == "" {
		option.FieldNameAccessor = defaultFieldNameAccessor
	}
	if option.MessageFormatter == nil {
		option.MessageFormatter = MessageFormatter{}
	}
	formatter := defaultMessageFormatter
	for k, v := range option.MessageFormatter {
		formatter[k] = v
	}
	_validator = &Validator{option.FieldNameAccessor, option.MessageFormatter}
}

// Validate validates the given struct and returns a map of errors
func Validate(d interface{}, options ...validator.Option) map[string]string {
	if _validator == nil {
		panic("validator is not initialized")
	}

	if reflect.ValueOf(d).Kind() != reflect.Struct {
		panic(fmt.Sprintf("Validate() only accepts structs; got %T", d))
	}

	t := reflect.TypeOf(d)
	validate := validator.New(options...)
	errs := validate.Struct(d)
	if errs == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range errs.(validator.ValidationErrors) {
		field, _ := t.FieldByName(err.Field())
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			log.Printf("json tag not found for field: %s", err.Field())
			fieldName = strings.ToLower(err.Field())
		}
		resolverMap, handled := _validator.messageFormatter[err.Tag()]
		if !handled {
			resolverMap = defaultFieldHandler
		}
		errors[fieldName] = resolverMap(&FieldError{FieldName: fieldName, FieldError: err})
	}
	return errors
}
