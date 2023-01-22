package generator

import (
	"encoding/json"
	"fmt"
	openapi "github.com/go-openapi/spec"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

// pathParamsProcessor parse path, search for path parameters. Create path for swagger annotation
func (s *SwaggerGenerator) pathParamsProcessor(op *openapi.Operation, path string) string {
	pParams := regexp.MustCompile(`:\w+`).FindAllString(path, -1)
	for _, param := range pParams {
		pName := param[1:]
		pathPlace := "{" + pName + "}"
		echoPathPlace := ":" + pName
		path = strings.ReplaceAll(path, echoPathPlace, pathPlace)
		sParam := openapi.Parameter{}
		sParam.Name = pName
		sParam.In = "path"
		sParam.Type = "string"
		op.Parameters = append(op.Parameters, sParam)
	}
	return path
}

func (s *SwaggerGenerator) queryParamsProcessor(op *openapi.Operation, paramType reflect.Type, skipParams map[string]struct{}) {
	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		fInfo := GetFieldInfo(field)
		if fInfo == nil {
			continue
		}
		if fInfo.In == "path" {
			continue
		}
		_, skip := skipParams[fInfo.Name]
		if skip {
			continue
		}
		if field.Type.Kind() != reflect.String {
			// Only type string supported for query parameters
			continue
		}
		sParam := openapi.Parameter{}
		sParam.Name = fInfo.Name
		sParam.In = "query"
		sParam.Type = "string"
		op.Parameters = append(op.Parameters, sParam)
	}
}

func (s *SwaggerGenerator) queryParamsOnlyProcessor(op *openapi.Operation, paramType reflect.Type) {
	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		fInfo := GetFieldInfo(field)
		if fInfo == nil {
			continue
		}
		if fInfo.In != "query" {
			continue
		}
		if field.Type.Kind() != reflect.String {
			// Only type string supported for query parameters
			continue
		}
		sParam := openapi.Parameter{}
		sParam.Name = fInfo.Name
		sParam.In = "query"
		sParam.Type = "string"
		op.Parameters = append(op.Parameters, sParam)
	}
}

func (s *SwaggerGenerator) bodyParamsProcessor(op *openapi.Operation, routeInfo RouteInfo) {
	if routeInfo.Handler.RequestType == nil {
		return
	}
	operationParams := make([]openapi.Parameter, 0)
	handlerInfo := routeInfo.Handler
	handlerInfo.FileUpload = append(handlerInfo.FileUpload, routeInfo.Parameters.FileUpload...)
	paramType := *routeInfo.Handler.RequestType
	pName := paramType.Name()
	s.definitionTypes[pName] = paramType
	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		if field.Type.Kind() == reflect.Struct {
			s.definitionTypes[field.Type.Name()] = field.Type
		}
	}
	opParam := s.generateSchemaBodyParam(pName, paramType)
	uploadParam := s.processFileUploadParam(paramType)
	for _, uParam := range uploadParam {
		var found bool
		for _, infoParam := range routeInfo.Handler.FileUpload {
			if uParam.Name == infoParam.Name {
				found = true
				break
			}
		}
		if !found {
			handlerInfo.FileUpload = append(handlerInfo.FileUpload, uParam)
		}
	}
	operationParams = append(op.Parameters, opParam)

	if len(handlerInfo.FileUpload) > 0 {
		defaultVal := reflect.New(paramType).Elem().Interface()
		structSkipFields, ok := s.defSkipFields[paramType.Name()]
		if !ok {
			structSkipFields = make([]string, 0)
		}
		for _, fileParam := range handlerInfo.FileUpload {
			var schemaFileParam *openapi.Parameter
			if fileParam.MultipleFiles {
				schemaFileParam = openapi.FormDataParam(fileParam.Name)
				schemaFileParam.Type = "array"
				schemaFileParam.Items = &openapi.Items{}
				schemaFileParam.Items.Typed("string", "binary")
			} else {
				schemaFileParam = openapi.FileParam(fileParam.Name)
			}
			op.Parameters = append(op.Parameters, *schemaFileParam)
			structSkipFields = append(structSkipFields, fileParam.jsonName)
		}

		// remove path, query, file parameters from default value for field
		dumpVal, err := json.Marshal(defaultVal)
		if err != nil {
			return
		}
		defaultMapVal := make(map[string]interface{})
		err = json.Unmarshal(dumpVal, &defaultMapVal)
		if err != nil {
			return
		}
		for _, fName := range structSkipFields {
			delete(defaultMapVal, fName)
		}
		dataParam := openapi.FormDataParam("request")
		dataParam.Default = defaultMapVal
		op.Parameters = append(op.Parameters, *dataParam)
	} else {
		op.Parameters = operationParams
	}
}

func (s *SwaggerGenerator) processFileUploadParam(paramType reflect.Type) []FileUploadParameters {
	fParams := make([]FileUploadParameters, 0)
	fHeaderType := reflect.TypeOf(&multipart.FileHeader{})
	s.definitionTypes[paramType.Name()] = paramType
	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		if field.Type == fHeaderType {
			fInfo := GetFieldInfo(field)
			if fInfo == nil {
				continue
			}
			fP := FileUploadParameters{
				Name:          fInfo.Name,
				jsonName:      fInfo.JSONName,
				MultipleFiles: false,
			}
			fParams = append(fParams, fP)
			continue
		} else if field.Type.Kind() == reflect.Slice {
			sElem := field.Type.Elem()
			if sElem == fHeaderType {
				fInfo := GetFieldInfo(field)
				if fInfo == nil {
					continue
				}
				fP := FileUploadParameters{
					Name:          fInfo.Name,
					jsonName:      fInfo.JSONName,
					MultipleFiles: true,
				}
				fParams = append(fParams, fP)
			}
		}
	}
	return fParams
}

func (s *SwaggerGenerator) responseProcessor(op *openapi.Operation, respType reflect.Type) {
	if respType == nil {
		return
	}
	op.Responses = &openapi.Responses{}
	op.Responses.StatusCodeResponses = make(map[int]openapi.Response)
	defName := getDefinitionName(respType)
	ref := openapi.ResponseRef(
		fmt.Sprintf("#/definitions/%s", defName),
	)
	respSchema := &openapi.Schema{
		SchemaProps: openapi.SchemaProps{
			Ref: ref.Ref,
		},
		ExtraProps: nil,
	}
	resp := &openapi.Response{}
	resp = resp.WithSchema(respSchema)
	op.Responses.StatusCodeResponses[http.StatusOK] = *resp

	s.definitionTypes[respType.Name()] = respType
	for i := 0; i < respType.NumField(); i++ {
		field := respType.Field(i)
		if field.Type.Kind() == reflect.Struct {
			if field.Type == timeType {
				continue
			}
			s.definitionTypes[field.Type.Name()] = field.Type
		}
	}
}

func (s *SwaggerGenerator) generateSchemaBodyParam(name string, reqType reflect.Type) openapi.Parameter {
	param := openapi.Parameter{}
	param.Name = name
	param.In = "body"
	param.Required = true
	defName := getDefinitionName(reqType)
	param.Schema = openapi.RefSchema(
		fmt.Sprintf("#/definitions/%s", defName),
	)
	return param
}

func GetFieldInfo(field reflect.StructField) *fieldInfo {
	res := &fieldInfo{}
	tagParts := strings.Split(field.Tag.Get("param"), ",")
	paramName := tagParts[0]
	if paramName == "-" {
		return nil
	}
	tagJsonParts := strings.Split(field.Tag.Get("json"), ",")
	jsonName := tagJsonParts[0]
	if jsonName == "-" {
		return nil
	}
	if len(jsonName) == 0 {
		jsonName = field.Name
	}
	if len(paramName) == 0 {
		res.Name = jsonName
		res.JSONName = jsonName
	} else {
		res.Name = paramName
		res.JSONName = jsonName
	}
	if len(tagParts) > 1 {
		res.In = tagParts[1]
	}
	return res
}
