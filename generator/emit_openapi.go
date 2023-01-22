package generator

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	openapi "github.com/go-openapi/spec"
)

type SwaggerGenerator struct {
	authTypes            map[string]AuthType
	defSkipFields        map[string][]string
	processedDefinitions map[string]struct{}
	definitionTypes      map[string]reflect.Type
}

var timeType = reflect.TypeOf(&time.Time{}).Elem()

func NewSwaggerGenerator() *SwaggerGenerator {
	return &SwaggerGenerator{
		defSkipFields:        make(map[string][]string),
		authTypes:            make(map[string]AuthType),
		processedDefinitions: make(map[string]struct{}),
		definitionTypes:      make(map[string]reflect.Type),
	}
}

func (s *SwaggerGenerator) EmitOpenAPIDefinition(routesMap map[string]RouteInfo) (openapi.Swagger, error) {
	sw := openapi.Swagger{}
	sw.Swagger = "2.0"
	sw.Info = &openapi.Info{}
	sw.Info.Version = "1.0"
	sw.Paths = &openapi.Paths{
		Paths: make(map[string]openapi.PathItem),
	}
	sw.Definitions = make(map[string]openapi.Schema)

	for path, routeInfo := range routesMap {
		pi := openapi.PathItem{}
		sPath := path
		switch routeInfo.Method {
		case http.MethodPost:
			sPath, pi.Post = s.processBodyParams(path, routeInfo)
		case http.MethodPatch:
			sPath, pi.Patch = s.processBodyParams(path, routeInfo)
		case http.MethodPut:
			sPath, pi.Put = s.processBodyParams(path, routeInfo)
		case http.MethodGet:
			sPath, pi.Get = s.processQueryParams(path, routeInfo)
		case http.MethodDelete:
			sPath, pi.Delete = s.processQueryParams(path, routeInfo)
		case http.MethodHead:
			sPath, pi.Head = s.processQueryParams(path, routeInfo)
		case http.MethodOptions:
			sPath, pi.Options = s.processQueryParams(path, routeInfo)
		}
		sw.Paths.Paths[sPath] = pi
	}
	secDefs, err := s.processSecurityDefinitions()
	if err != nil {
		return openapi.Swagger{}, fmt.Errorf("cannot process security definition: %w", err)
	}
	sw.SecurityDefinitions = secDefs
	sw.Definitions = s.processDefinitions(s.definitionTypes)

	return sw, nil
}

func (s *SwaggerGenerator) processSecurityDefinitions() (openapi.SecurityDefinitions, error) {
	secDefs := openapi.SecurityDefinitions{}
	for _, aType := range s.authTypes {
		var authScheme *openapi.SecurityScheme
		if aType.BasicAuth != nil {
			authScheme = &openapi.SecurityScheme{
				SecuritySchemeProps: openapi.SecuritySchemeProps{
					Description: aType.Description,
					Type:        BasicAuth,
				},
			}
		} else if aType.APIKey != nil {
			authScheme = &openapi.SecurityScheme{
				SecuritySchemeProps: openapi.SecuritySchemeProps{
					Description: aType.Description,
					Type:        APIKey,
					Name:        aType.APIKey.Name,
					In:          aType.APIKey.In,
				},
			}
		} else if aType.OAuth2 != nil {
			authScheme = &openapi.SecurityScheme{
				SecuritySchemeProps: openapi.SecuritySchemeProps{
					Description:      aType.Description,
					Type:             OAuth2,
					Flow:             aType.OAuth2.Flow,
					AuthorizationURL: aType.OAuth2.AuthorizationURL,
					TokenURL:         aType.OAuth2.TokenURL,
					Scopes:           aType.OAuth2.Scopes,
				},
			}
		}
		if authScheme == nil {
			return openapi.SecurityDefinitions{}, errors.New("auth scheme not defined")
		}
		secDefs[aType.AuthTypeName] = authScheme

	}
	return secDefs, nil
}

func (s *SwaggerGenerator) processAuthParams(op *openapi.Operation, params HandlerParameters) {
	for _, authParam := range params.Auth {
		if authParam.Scopes == nil {
			authParam.Scopes = []string{}
		}
		s.authTypes[authParam.AuthTypeName] = authParam
		authVaL := map[string][]string{authParam.AuthTypeName: authParam.Scopes}
		op.Security = []map[string][]string{authVaL}
	}
}

// processQueryParams process request struct as query parameters struct. Only skip fields defined In skipParams
func (s *SwaggerGenerator) processQueryParams(path string, routeInfo RouteInfo) (string, *openapi.Operation) {
	op := &openapi.Operation{}
	op.Tags = append(op.Tags, routeInfo.Tags...)
	op.Summary = routeInfo.Parameters.Summary
	sPath := s.pathParamsProcessor(op, path)
	skipParams := make(map[string]struct{})
	for _, pParam := range op.Parameters {
		skipParams[pParam.Name] = struct{}{}
	}
	paramType := routeInfo.Handler.RequestType
	if paramType != nil {
		s.queryParamsProcessor(op, *paramType, skipParams)
	}
	s.processAuthParams(op, routeInfo.Parameters)
	if routeInfo.Handler.OutputType != nil {
		s.responseProcessor(op, *routeInfo.Handler.OutputType)
	}

	return sPath, op
}

func (s *SwaggerGenerator) processBodyParams(path string, routeInfo RouteInfo) (string, *openapi.Operation) {
	op := &openapi.Operation{}
	op.Tags = append(op.Tags, routeInfo.Tags...)
	op.Summary = routeInfo.Parameters.Summary
	sPath := s.pathParamsProcessor(op, path)
	if routeInfo.Handler.RequestType != nil {
		reqType := *routeInfo.Handler.RequestType
		s.queryParamsOnlyProcessor(op, reqType)
		skipParams := make([]string, 0, len(op.Parameters))
		for _, pParam := range op.Parameters {
			skipParams = append(skipParams, pParam.Name)
		}
		paramName := reqType.Name()
		s.defSkipFields[paramName] = skipParams
		s.bodyParamsProcessor(op, routeInfo)

	}
	s.processAuthParams(op, routeInfo.Parameters)
	if routeInfo.Handler.OutputType != nil {
		s.responseProcessor(op, *routeInfo.Handler.OutputType)
	}
	return sPath, op
}

func (s *SwaggerGenerator) processDefinitions(definitionTypes map[string]reflect.Type) openapi.Definitions {
	referencedDefinitions := make(map[string]reflect.Type)
	definitions := make(openapi.Definitions)
	for definitionName, definitionType := range definitionTypes {
		_, ok := s.processedDefinitions[definitionName]
		if ok {
			continue
		}
		defSkipFields, ok := s.defSkipFields[definitionName]
		if !ok {
			defSkipFields = make([]string, 0)
		}
		props := make(map[string]openapi.Schema)
		for i := 0; i < definitionType.NumField(); i++ {
			field := definitionType.Field(i)
			fInfo := GetFieldInfo(field)
			fieldName := fInfo.Name
			var skip bool
			for _, fName := range defSkipFields {
				if fName == fieldName {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			schema, addition := getSchemaType(field.Type)
			if addition != nil {
				for defName, defType := range addition {
					referencedDefinitions[defName] = defType
				}
			}
			if schema == nil {
				continue
			}
			props[fieldName] = *schema
		}

		var definition openapi.Schema
		definition.Type = []string{"object"}
		definition.Properties = props
		defName := getDefinitionName(definitionType)
		definitions[defName] = definition
		s.processedDefinitions[definitionName] = struct{}{}
	}
	if len(referencedDefinitions) > 0 {
		refDefs := s.processReferencedDefinitions(referencedDefinitions)
		for defName, defRes := range refDefs {
			definitions[defName] = defRes
			s.processedDefinitions[defName] = struct{}{}
		}
	}
	return definitions
}

func (s *SwaggerGenerator) processReferencedDefinitions(definitionTypes map[string]reflect.Type) openapi.Definitions {
	definitions := make(openapi.Definitions)
	structDefinitions := make(map[string]reflect.Type)
	for definitionName, definitionType := range definitionTypes {
		_, ok := s.processedDefinitions[definitionName]
		if ok {
			continue
		}
		defKind := definitionType.Kind()
		defName := getDefinitionName(definitionType)
		switch defKind {
		case reflect.Struct:
			structDefinitions[defName] = definitionType
		default:
			var definition openapi.Schema
			definition.Type = processNonStructDefinitionType(defKind)
			definitions[defName] = definition
		}
	}
	if len(structDefinitions) > 0 {
		structDefs := s.processDefinitions(structDefinitions)
		for defName, definition := range structDefs {
			definitions[defName] = definition
		}
	}
	return definitions
}

func processNonStructDefinitionType(defKind reflect.Kind) []string {
	switch defKind {
	case reflect.String:
		return []string{"string"}
	case reflect.Float32, reflect.Float64:
		return []string{"number"}
	case reflect.Bool:
		return []string{"boolean"}
	default:
		if reflect.Int <= defKind && defKind <= reflect.Uint64 {
			return []string{"integer"}
		}
	}
	return []string{}
}

func getSchemaType(paramType reflect.Type) (*openapi.Schema, map[string]reflect.Type) {
	switch paramType.Kind() {
	case reflect.Bool:
		return openapi.BoolProperty(), nil
	case reflect.Int8:
		return openapi.Int8Property(), nil
	case reflect.Int16:
		return openapi.Int16Property(), nil
	case reflect.Int32:
		return openapi.Int32Property(), nil
	case reflect.Int, reflect.Int64:
		return openapi.Int64Property(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return openapi.Int64Property(), nil
	case reflect.Float32:
		return openapi.Float32Property(), nil
	case reflect.Float64:
		return openapi.Float64Property(), nil
	case reflect.String:
		return openapi.StringProperty(), nil
	case reflect.Slice:
		fieldType, additionalDefinition := getSchemaType(paramType.Elem())
		return openapi.ArrayProperty(fieldType), additionalDefinition
	case reflect.Array:
		fieldType, additionalDefinition := getSchemaType(paramType.Elem())
		return openapi.ArrayProperty(fieldType), additionalDefinition
	case reflect.Map:
		fieldType, additionalDefinition := getSchemaType(paramType.Elem())
		return openapi.MapProperty(fieldType), additionalDefinition
	case reflect.Struct:
		if paramType == timeType {
			return openapi.DateTimeProperty(), nil
		}
		addition := map[string]reflect.Type{paramType.Name(): paramType}
		defName := getDefinitionName(paramType)
		return openapi.RefProperty(
			fmt.Sprintf("#/definitions/%s", defName),
		), addition
	case reflect.Ptr:
		refVal := reflect.New(paramType).Elem()
		refElem := refVal.Type().Elem()
		if refElem == timeType {
			return openapi.DateTimeProperty(), nil
		}
		defName := getDefinitionName(refElem)
		resType := openapi.RefProperty(
			fmt.Sprintf("#/definitions/%s", defName),
		)
		addition := map[string]reflect.Type{refElem.Name(): refElem}
		return resType, addition
	case reflect.Interface:
		return &openapi.Schema{
			SchemaProps: openapi.SchemaProps{
				AnyOf: []openapi.Schema{
					*openapi.StringProperty(),
					{SchemaProps: openapi.SchemaProps{Type: []string{"integer"}}},
					{SchemaProps: openapi.SchemaProps{Type: []string{"number"}}},
					*openapi.BoolProperty(),
				}},
			SwaggerSchemaProps: openapi.SwaggerSchemaProps{Example: "any_value"},
		}, nil
	}
	return nil, nil
}
