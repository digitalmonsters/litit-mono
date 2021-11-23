package swagger

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/digitalmonsters/go-common/common"
	"github.com/rs/zerolog/log"
	"reflect"
	"strings"
)

type IApiCommand interface {
	GetPath() string
	RequireIdentityValidation() bool
	AccessLevel() common.AccessLevel
	GetHttpMethod() string
}

type ApiDescription struct {
	Request           interface{}
	Response          interface{}
	MethodDescription string
	Summary           string
	Tags              []string
}

type ConstantDescription struct {
	Ref    interface{}
	Values []string
}

type constantMapItem struct {
	Example string
	Values  []string
}

var constantMap = map[string]*constantMapItem{}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GenerateDoc(cmd []IApiCommand, apiDescriptions map[string]ApiDescription, constants []ConstantDescription) map[string]interface{} {
	for key, val := range apiDescriptions { // normalize
		k := strings.ToLower(key)

		if key != k {
			apiDescriptions[k] = val
			delete(apiDescriptions, key)
		}
	}

	authScopes := map[string]string{}
	root := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"description": "api spec",
			"version":     "0.0.1",
			"title":       "api spec",
		},
		"securityDefinitions": map[string]interface{}{
			jwtAuthName: map[string]interface{}{
				"type":             "oauth2",
				"flow":             "implicit",
				"authorizationUrl": "http://petstore.swagger.io/oauth/dialog",
				"scopes":           authScopes,
			},
		},
	}

	if len(constants) > 0 {
		for _, c := range constants {
			realType := reflect.TypeOf(c.Ref)

			if !isSimple(realType.Kind()) {
				continue
			}

			name := getTypeName(realType)

			if _, ok := constantMap[name]; !ok {
				constantMap[name] = &constantMapItem{
					Example: fmt.Sprint(c.Ref),
				}
			}

			for _, v := range c.Values {
				if !contains(constantMap[name].Values, v) {
					constantMap[name].Values = append(constantMap[name].Values, v)
				}
			}
		}
	}

	allDefs := prepareDefs(apiDescriptions)

	root["definitions"] = buildDefs(allDefs)
	root["paths"] = buildPath(cmd, apiDescriptions, authScopes)

	return root
}

func mapInnerKind(kind reflect.Kind, name string) string {
	if v, ok := customTypesLogic[name]; ok {
		return v.SwaggerType
	}

	return mapKindToSwaggerType(kind)
}

const jwtAuthName = "jwt"

func buildPath(commands []IApiCommand, apiDescriptions map[string]ApiDescription, scopes map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, cmd := range commands {
		originalName := strings.ToLower(cmd.GetPath())
		pathName := originalName
		if !strings.HasPrefix(pathName, "/") {
			pathName = fmt.Sprintf("/%v", pathName)
		}

		methodInfo := map[string]interface{}{
			"description": cmd.GetPath(),
			"summary":     cmd.GetPath(),
			"produces":    []string{"application/json"},
			"consumes":    []string{"application/json"},
			"operationId": pathName,
			"parameters":  []interface{}{},
			"responses":   map[string]interface{}{},
		}

		if cmd.AccessLevel() > common.AccessLevelPublic || cmd.RequireIdentityValidation() {
			permissionName := ""
			if cmd.AccessLevel() == common.AccessLevelPublic {
				permissionName = "global:authorized"
			} else {
				permissionName = fmt.Sprintf("%v:%v", cmd.AccessLevel().ToString(), strings.ToLower(cmd.GetPath()))
			}

			scopes[permissionName] = permissionName

			methodInfo["security"] = []interface{}{
				map[string][]interface{}{
					jwtAuthName: []interface{}{
						permissionName,
					},
				},
			}
		}

		hasResponse := false

		if notV, ok := apiDescriptions[originalName]; ok {
			if len(notV.MethodDescription) > 0 {
				methodInfo["description"] = notV.MethodDescription
			}

			if len(notV.Summary) > 0 {
				methodInfo["summary"] = notV.Summary
			}

			if len(notV.Tags) > 0 {
				methodInfo["tags"] = notV.Tags
			}
			for i, item := range []interface{}{notV.Request, notV.Response} {
				paramsMap := map[string]interface{}{}

				isRequest := i == 0
				topType := skipPointers(reflect.TypeOf(item))

				if topType == nil {
					continue
				}

				if isRequest {
					paramsMap["in"] = "body"
					paramsMap["name"] = "body"
					//paramsMap["description"] = "some dec2"
					paramsMap["required"] = true
				} else {
					paramsMap["description"] = "sample response"
				}

				swType := mapKindToSwaggerType(topType.Kind())

				name, _, innerType := getRealType(reflect.TypeOf(item))
				simple := isSimple(innerType.Kind())

				var customDef *customTypeDef
				if c, ok1 := customTypesLogic[name]; ok1 {
					simple = true
					customDef = &c
				}

				if swType == "array" {
					if simple {
						innerSw := mapInnerKind(innerType.Kind(), name)

						itemsMap := map[string]interface{}{
							"type": innerSw,
						}

						if customDef != nil && len(customDef.SwaggerFormat) > 0 {
							itemsMap["format"] = customDef.SwaggerFormat
						}

						if con, ok1 := constantMap[name]; ok1 {
							itemsMap["enum"] = con.Values
							itemsMap["example"] = con.Example
						}

						paramsMap["schema"] = map[string]interface{}{
							"type":  swType,
							"items": itemsMap,
						}
					} else {
						paramsMap["schema"] = map[string]interface{}{
							"type": swType,
							"items": map[string]interface{}{
								"$ref": fmt.Sprintf("#/definitions/%v", name),
							},
						}
					}
				} else if swType == "map" {
					if simple {
						innerSw := mapInnerKind(innerType.Kind(), name)

						itemsMap := map[string]interface{}{
							"type": innerSw,
						}

						if customDef != nil {
							itemsMap["nullable"] = customDef.IsNullable
						}

						if con, ok1 := constantMap[name]; ok1 {
							itemsMap["enum"] = con.Values
							itemsMap["example"] = con.Example
						}

						if customDef != nil && len(customDef.SwaggerFormat) > 0 {
							itemsMap["format"] = customDef.SwaggerFormat
						}

						paramsMap["schema"] = map[string]interface{}{
							"type":                 "object",
							"additionalProperties": itemsMap,
						}
					} else {
						paramsMap["schema"] = map[string]interface{}{
							"type": "object",
							"additionalProperties": map[string]interface{}{
								"$ref": fmt.Sprintf("#/definitions/%v", name),
							},
						}
					}
				} else {
					if simple {
						innerSw := mapInnerKind(innerType.Kind(), name)

						itemsMap := map[string]interface{}{
							"type": innerSw,
						}

						if customDef != nil {
							itemsMap["nullable"] = customDef.IsNullable
						}

						if con, ok1 := constantMap[name]; ok1 {
							itemsMap["enum"] = con.Values
							itemsMap["example"] = con.Example
						}

						if name == "time_Time" {
							itemsMap["format"] = "date-time"
						}

						paramsMap["schema"] = itemsMap

					} else {
						paramsMap["schema"] = map[string]interface{}{
							"$ref": fmt.Sprintf("#/definitions/%v", name),
						}
					}
				}

				if isRequest {
					methodInfo["parameters"] = []interface{}{paramsMap}
				} else {
					methodInfo["responses"] = map[string]interface{}{
						"200": paramsMap,
					}

					hasResponse = true
				}
			}
		}

		if !hasResponse {
			methodInfo["responses"] = map[string]interface{}{
				"200": map[string]interface{}{
					"description": "response",
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			}
		}

		result[pathName] = map[string]interface{}{
			strings.ToLower(cmd.GetHttpMethod()): methodInfo,
		}
	}

	return result
}

func appendDefsToResult(def reflect.Type, innerProps map[string]interface{}, allDefs map[string]reflect.Type, required *[]string) {
	for i := 0; i < def.NumField(); i++ {
		swaggerFieldDefinition := map[string]interface{}{}

		field := def.Field(i)

		if field.Anonymous {
			embeddedTypeName := getTypeName(field.Type)

			if v, ok := allDefs[embeddedTypeName]; ok {
				appendDefsToResult(v, innerProps, allDefs, required)
				continue
			}
		}

		fieldName := field.Name

		if js := field.Tag.Get("json"); len(js) > 0 {
			fieldName = js
		}

		if swag := field.Tag.Get("swag"); len(swag) > 0 {
			for _, group := range strings.Split(swag, ";") {
				if len(group) == 0 {
					continue
				}

				parsedGroup := strings.Split(group, ":")

				if parsedGroup[0] == "required" {
					*required = append(*required, fieldName)
				}
			}
		}

		topType := skipPointers(field.Type)
		swType := mapKindToSwaggerType(topType.Kind())

		swaggerFieldDefinition["type"] = swType

		if swType == "map" {
			swaggerFieldDefinition["type"] = "object"
		}

		name, _, innerType := getRealType(field.Type)

		simple := isSimple(innerType.Kind())

		var customType *customTypeDef
		if c, ok1 := customTypesLogic[name]; ok1 {
			simple = true
			customType = &c
		}

		if swType == "array" {
			if simple {
				innerSw := mapInnerKind(innerType.Kind(), name)

				itemsMap := map[string]interface{}{
					"type": innerSw,
				}

				if customType != nil {
					if len(customType.SwaggerFormat) > 0 {
						itemsMap["format"] = "date-time"
					}
					itemsMap["nullable"] = customType.IsNullable
				}

				if con, ok1 := constantMap[name]; ok1 {
					itemsMap["enum"] = con.Values
					itemsMap["example"] = con.Example
				}

				swaggerFieldDefinition["items"] = itemsMap
			} else {
				swaggerFieldDefinition["items"] = map[string]interface{}{
					"$ref": fmt.Sprintf("#/definitions/%v", name),
				}
			}
		} else if swType == "map" {
			if simple {
				innerSw := mapInnerKind(innerType.Kind(), name)

				itemsMap := map[string]interface{}{
					"type": innerSw,
				}
				if customType != nil {
					if len(customType.SwaggerFormat) > 0 {
						itemsMap["format"] = "date-time"
					}
					itemsMap["nullable"] = customType.IsNullable
				}

				if con, ok1 := constantMap[name]; ok1 {
					itemsMap["enum"] = con.Values
					itemsMap["example"] = con.Example
				}

				swaggerFieldDefinition["additionalProperties"] = itemsMap
			} else {
				swaggerFieldDefinition["additionalProperties"] = map[string]interface{}{
					"$ref": fmt.Sprintf("#/definitions/%v", name),
				}
			}
		} else {
			if simple {
				innerSw := mapInnerKind(innerType.Kind(), name)

				if customType != nil {
					if len(customType.SwaggerFormat) > 0 {
						swaggerFieldDefinition["format"] = "date-time"
					}
					swaggerFieldDefinition["nullable"] = customType.IsNullable
				}

				if con, ok1 := constantMap[name]; ok1 {
					swaggerFieldDefinition["enum"] = con.Values
					swaggerFieldDefinition["example"] = con.Example
				}

				swaggerFieldDefinition["type"] = innerSw
			} else {
				swaggerFieldDefinition["$ref"] = fmt.Sprintf("#/definitions/%v", name)
			}
		}

		innerProps[fieldName] = swaggerFieldDefinition
	}
}

func buildDefs(allDefs map[string]reflect.Type) map[string]interface{} {
	result := make(map[string]interface{})

	for key, def := range allDefs {
		topProps := map[string]interface{}{
			"type": "object",
		}

		innerProps := map[string]interface{}{}
		topProps["properties"] = innerProps

		result[key] = topProps

		var required []string

		appendDefsToResult(def, innerProps, allDefs, &required)

		if len(required) > 0 {
			topProps["required"] = required
		}
	}

	return result
}

func mapKindToSwaggerType(kind reflect.Kind) string {
	swType := ""

	switch kind {
	case reflect.Map:
		swType = "map"
	case reflect.Struct:
		swType = "object"
	case reflect.Interface:
		swType = "object"
	case reflect.Slice:
		swType = "array"
	case reflect.Bool:
		swType = "boolean"
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uintptr:
		swType = "integer"
	case reflect.Uint64:
		fallthrough
	case reflect.Int64:
		swType = "long"
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		swType = "number"
	case reflect.String:
		swType = "string"
	default:
		log.Logger.Warn().Msg(fmt.Sprintf("can not map type %v", spew.Sdump(kind)))
	}

	return swType
}

func prepareDefs(apiDescriptions map[string]ApiDescription) map[string]reflect.Type {
	toGenerate := map[string]reflect.Type{}

	for _, i := range apiDescriptions {
		for _, inner := range []interface{}{i.Request, i.Response} {
			if inner == nil {
				continue
			}

			name, _, realTypeDef := getRealType(reflect.TypeOf(inner))

			if name == "time_Time" {
				continue
			}

			if isSimple(realTypeDef.Kind()) {
				continue
			}

			if _, ok := toGenerate[name]; !ok {
				toGenerate[name] = realTypeDef
			}
		}
	}

	hasNewRecords := false
	for {
		for _, item := range toGenerate {
			for i := 0; i < item.NumField(); i++ {
				fieldData := item.Field(i)
				name, _, realTypeDef := getRealType(skipPointers(fieldData.Type))

				if name == "time_Time" {
					continue
				}

				if isSimple(realTypeDef.Kind()) {
					continue
				}

				if _, ok := toGenerate[name]; !ok {
					toGenerate[name] = realTypeDef
					hasNewRecords = true
				}
			}
		}

		if hasNewRecords {
			hasNewRecords = false
			continue
		}

		break
	}

	return toGenerate
}

func isSimple(kind reflect.Kind) bool {
	if kind != reflect.Ptr && kind != reflect.Slice && kind != reflect.Map && kind != reflect.Struct &&
		kind != reflect.Invalid {
		return true
	}

	return false
}

func skipPointers(expectedToBeStruct reflect.Type) reflect.Type {
	if expectedToBeStruct == nil {
		return nil
	}

	if expectedToBeStruct.Kind() == reflect.Ptr {
		n := skipPointers(expectedToBeStruct.Elem())
		return n
	}

	return expectedToBeStruct
}

type customTypeDef struct {
	SwaggerType   string
	SwaggerFormat string
	IsNullable    bool
}

var customTypesLogic = map[string]customTypeDef{
	"github.com.shopspring.decimal_Decimal":     customTypeDef{SwaggerType: "number"},
	"github.com.shopspring.decimal_NullDecimal": customTypeDef{SwaggerType: "number", IsNullable: true},
	"gopkg.in.guregu.null.v4_Time":              customTypeDef{SwaggerType: "string", SwaggerFormat: "date-time", IsNullable: true},
	"gopkg.in.guregu.null.v4_String":            customTypeDef{SwaggerType: "string", IsNullable: true},
	"gopkg.in.guregu.null.v4_Int":               customTypeDef{SwaggerType: "integer", IsNullable: true},
	"gopkg.in.guregu.null.v3_Time":              customTypeDef{SwaggerType: "string", SwaggerFormat: "date-time", IsNullable: true},
	"gopkg.in.guregu.null.v4_Bool":              customTypeDef{SwaggerType: "boolean", IsNullable: true},
	"database.sql_NullInt64":                    customTypeDef{SwaggerType: "long", IsNullable: true},
	"database.sql_NullTime":                     customTypeDef{SwaggerType: "string", SwaggerFormat: "date-time", IsNullable: true},
	"math.big_Int":                              customTypeDef{SwaggerType: "long"},
	"time_Time":                                 customTypeDef{SwaggerType: "string", SwaggerFormat: "date-time"},
}

func getRealType(expectedToBeStruct reflect.Type) (string, reflect.Kind, reflect.Type) {
	val := expectedToBeStruct

	typeName := getTypeName(val)
	_, isCustom := customTypesLogic[typeName]

	if val.Kind() == reflect.Struct && isCustom {
		return getTypeName(val), reflect.Invalid, val
	}

	if val.Kind() == reflect.Slice {
		n, _, t := getRealType(expectedToBeStruct.Elem())
		return n, val.Kind(), t
	}
	if val.Kind() == reflect.Ptr {
		n, _, t := getRealType(expectedToBeStruct.Elem())
		return n, val.Kind(), t
	}
	if val.Kind() == reflect.Map {
		n, _, t := getRealType(expectedToBeStruct.Elem())
		return n, val.Kind(), t
	}

	return getTypeName(val), val.Kind(), val
}

var replacementArr = []string{"/", "-"}

func getTypeName(val reflect.Type) string {
	result := fmt.Sprintf("%v_%v", val.PkgPath(), val.Name())

	for _, i := range replacementArr {
		result = strings.ReplaceAll(result, i, ".")
	}

	return result
}
