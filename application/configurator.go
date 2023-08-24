package application

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type Configurator[T any] struct {
	builder ConfiguratorBuilder[T]
	raw     map[string]string
	keys    []string
	Values  T
}

func (c *Configurator[T]) GetRawData() map[string]string {
	return c.raw
}

func (c *Configurator[T]) init() {
	var updated T

	typeData := reflect.TypeOf(updated)

	for i := 0; i < typeData.NumField(); i++ {
		fieldData := typeData.Field(i)

		c.keys = append(c.keys, fieldData.Name)
	}
}

func (c *Configurator[T]) setValues(inputData map[string]string) error {
	var updated T

	var err error

	typeData := reflect.TypeOf(updated)

	for i := 0; i < typeData.NumField(); i++ {
		fieldData := typeData.Field(i)

		inputField, ok := inputData[fieldData.Name]

		if !ok {
			err = multierror.Append(err, errors.New(fmt.Sprintf("key [%v] not found", fieldData.Name)))

			continue
		}

		if parseErr := c.setValue(&updated, fieldData.Name, inputField); parseErr != nil {
			err = multierror.Append(err, parseErr)
		}
	}

	if err != nil {
		return err
	}

	c.Values = updated
	c.raw = inputData

	return err
}

func (c Configurator[T]) setValue(instance *T, key string, value string) error {
	typeData := reflect.TypeOf(instance).Elem()

	for i := 0; i < typeData.NumField(); i++ {
		fieldData := typeData.Field(i)

		if key == fieldData.Name {
			fieldSetData := reflect.ValueOf(instance).Elem().Field(i)
			if !fieldSetData.CanSet() {
				return errors.New(fmt.Sprintf("field %v can not be set by reflection", fieldData.Name))
			}

			if parsedValue, err := c.parseValue(fieldData, value); err != nil {
				return err
			} else {
				fieldSetData.Set(parsedValue)
			}
		}
	}

	return nil
}

func (c Configurator[T]) parseValue(fieldData reflect.StructField, value string) (reflect.Value, error) {
	typeName := fieldData.Type.String()

	switch typeName {
	case "decimal.Decimal":
		if parsed, err := decimal.NewFromString(value); err != nil {
			return reflect.Value{}, errors.New(fmt.Sprintf("field %v can not be parsed to %v because of error: %v",
				fieldData.Name, typeName, err.Error()))
		} else {
			return reflect.ValueOf(parsed), nil
		}
	case "bool":
		if parsed, err := strconv.ParseBool(value); err != nil {
			return reflect.Value{}, errors.WithStack(err)
		} else {
			return reflect.ValueOf(parsed), nil
		}
	case "string":
		return reflect.ValueOf(value), nil
	case "int64":
		fallthrough
	case "int":
		if parsed, err := strconv.ParseInt(value, 10, 64); err != nil {
			return reflect.Value{}, errors.WithStack(err)
		} else {
			if typeName == "int" {
				return reflect.ValueOf(int(parsed)), nil
			}
			return reflect.ValueOf(parsed), nil
		}
	default:
		return reflect.Value{}, errors.New(fmt.Sprintf("field %v has unsupported type by parser %v", fieldData.Name, typeName))
	}
}

func (c *Configurator[T]) Refresh(ctx context.Context) error {
	values, err := c.builder.retriever.Retrieve(c.keys, ctx)
	if err != nil {
		return fmt.Errorf("refresh err: %s", err.Error())
	}

	return c.setValues(values)
}
