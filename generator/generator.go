package generator

import (
	"fmt"
	openapi "github.com/go-openapi/spec"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"time"
	"unicode"
)

const definitionPrefix = "#/definitions/"

func float64Ptr(f float64) *float64 {
	return &f
}

func unsignedType(schema *openapi.Schema) *openapi.Schema {
	schema.SchemaProps.Minimum = float64Ptr(0)
	return schema
}

type SchemaGenerator struct {
	processTypes []reflect.Type
	cTypes       map[reflect.Type]*openapi.Schema
}

var simpleTypesMapping = map[reflect.Kind]*openapi.Schema{
	reflect.Bool:    openapi.BoolProperty(),
	reflect.Int:     openapi.Int64Property(),
	reflect.Int8:    openapi.Int8Property(),
	reflect.Int16:   openapi.Int16Property(),
	reflect.Int32:   openapi.Int32Property(),
	reflect.Int64:   openapi.Int64Property(),
	reflect.Uint:    unsignedType(openapi.Int64Property()),
	reflect.Uint8:   unsignedType(openapi.Int8Property()),
	reflect.Uint16:  unsignedType(openapi.Int16Property()),
	reflect.Uint32:  unsignedType(openapi.Int32Property()),
	reflect.Uint64:  unsignedType(openapi.Int64Property()),
	reflect.Float32: openapi.Float32Property(),
	reflect.Float64: openapi.Float64Property(),
	reflect.String:  openapi.StringProperty(),
}

func NewSchemaGenerator() *SchemaGenerator {
	timeType := reflect.TypeOf(&time.Time{}).Elem()
	uuidType := reflect.TypeOf(&uuid.UUID{}).Elem()
	return &SchemaGenerator{
		processTypes: make([]reflect.Type, 0),
		cTypes: map[reflect.Type]*openapi.Schema{
			timeType: openapi.DateTimeProperty(),
			uuidType: UUIDProperty(),
		},
	}
}

func (gen *SchemaGenerator) GetSchema(paramType reflect.Type) (openapi.Definitions, error) {
	defs := openapi.Definitions{}
	gen.processTypes = append(gen.processTypes, paramType)
	for i := 0; i < len(gen.processTypes); i++ {
		param := gen.processTypes[i]
		if param.Kind() != reflect.Struct {
			// can generate schema only for structs
			continue
		}
		defStructName := getDefinitionName(param)
		structSchema, err := gen.processStruct(param)
		if err != nil {
			return nil, fmt.Errorf("cannot process parameter %s: %w", defStructName, err)
		}
		defName := definitionPrefix + defStructName
		defs[defName] = *structSchema
	}
	return defs, nil
}

func (gen *SchemaGenerator) processParam(paramType reflect.Type) (*openapi.Schema, error) {
	paramKind := paramType.Kind()
	simpleType, ok := simpleTypesMapping[paramKind]
	if ok {
		return simpleType, nil
	}
	switch paramType.Kind() {
	case reflect.Slice:
		cType := gen.tryCustomType(paramType)
		if cType != nil {
			return cType, nil
		}
		pTypeElem := paramType.Elem()
		fieldType, additionalDefinition := gen.processParam(pTypeElem)
		return openapi.ArrayProperty(fieldType), additionalDefinition
	case reflect.Array:
		cType := gen.tryCustomType(paramType)
		if cType != nil {
			return cType, nil
		}
		fieldType, err := gen.processParam(paramType.Elem())
		if err != nil {
			return nil, fmt.Errorf("cannot process array property: %w", err)
		}
		return openapi.ArrayProperty(fieldType), nil
	case reflect.Map:
		cType := gen.tryCustomType(paramType)
		if cType != nil {
			return cType, nil
		}
		fieldType, err := gen.processParam(paramType.Elem())
		if err != nil {
			return nil, fmt.Errorf("cannot process map property: %w", err)
		}
		return openapi.MapProperty(fieldType), nil
	case reflect.Struct:
		cType := gen.tryCustomType(paramType)
		if cType != nil {
			return cType, nil
		}
		defName := getDefinitionName(paramType)
		gen.processTypes = append(gen.processTypes, paramType)
		return openapi.RefProperty(
			definitionPrefix + defName,
		), nil
	case reflect.Ptr:
		refElem := reflect.New(paramType).Elem().Type().Elem()
		eName := refElem.Name()
		_ = eName
		return gen.processParam(refElem)
	case reflect.Interface:
		return &openapi.Schema{
			SchemaProps: openapi.SchemaProps{
				AnyOf: []openapi.Schema{
					*openapi.StringProperty(),
					{SchemaProps: openapi.SchemaProps{Type: []string{"integer"}}},
					{SchemaProps: openapi.SchemaProps{Type: []string{"number"}}},
					*openapi.BoolProperty(),
				}},
			SwaggerSchemaProps: openapi.SwaggerSchemaProps{Example: "any value"},
		}, nil
	}
	return nil, nil
}

func (gen *SchemaGenerator) processStruct(paramType reflect.Type) (*openapi.Schema, error) {
	res := &openapi.Schema{}
	res.Type = []string{"object"}
	res.Properties = openapi.SchemaProperties{}

	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		if unicode.IsLower([]rune(field.Name)[0]) {
			continue
		}
		fieldName := field.Tag.Get("json")
		if fieldName == "-" {
			continue
		}
		if fieldName == "" {
			fieldName = field.Name
		}

		schema, err := gen.processParam(field.Type)
		if err != nil {
			return nil, fmt.Errorf("cannot process field %s: %w", fieldName, err)
		}
		if schema == nil {
			// Pass unsupported types
			continue
		}
		res.Properties[fieldName] = *schema
	}
	return res, nil
}

func (gen *SchemaGenerator) tryCustomType(paramType reflect.Type) *openapi.Schema {
	cSchema, ok := gen.cTypes[paramType]
	if ok {
		return cSchema
	}
	if paramType.Kind() != reflect.Struct {
		return nil
	}
	if paramType.NumField() != 1 {
		return nil
	}
	sField := paramType.Field(0)
	if !sField.Anonymous {
		return nil
	}
	cSchema, ok = gen.cTypes[sField.Type]
	if ok {
		return cSchema
	}

	return nil
}

func UUIDProperty() *openapi.Schema {
	fType := openapi.StringProperty()
	fType.SchemaProps.Format = "uuid"
	return fType
}

func getDefinitionName(defType reflect.Type) string {
	pParts := strings.Split(defType.PkgPath(), "/")
	lastPart := pParts[len(pParts)-1]
	if len(lastPart) > 0 {
		lastPart += "."
	}
	defName := lastPart + defType.Name()
	return defName
}
