package wrapper

import (
	"encoding/json"
	"fmt"
	"github.com/AlhimicMan/goswag/generator"
	"github.com/labstack/echo/v4"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
)

type ReqField struct {
	ParamName       string
	StructFieldName string
}

type fileField struct {
	fieldName       string
	structFieldName string
	multiple        bool
}

var fHeaderType = reflect.TypeOf(&multipart.FileHeader{})

func (g *WrapGroup) callProcessor(path string, handler interface{}, processBody bool) echo.HandlerFunc {
	handlerType := reflect.TypeOf(handler)
	inParamsCount := handlerType.NumIn()
	outParamsCount := handlerType.NumOut()
	reqParam := handlerType.In(1)
	pathParamNames := make([]string, 0)
	pParams := regexp.MustCompile(`:\w+`).FindAllString(path, -1)
	for _, param := range pParams {
		pName := param[1:]
		pathParamNames = append(pathParamNames, pName)
	}
	queryParams, pathParams := g.getParams(reqParam, pathParamNames, processBody)
	fParams := g.getUploadFileParams(reqParam)
	handlerFunc := reflect.ValueOf(handler)
	return func(c echo.Context) error {
		inputVal := reflect.New(reqParam)
		if processBody {
			if len(fParams) == 0 {
				inputValPtr := inputVal.Interface()
				err := json.NewDecoder(c.Request().Body).Decode(inputValPtr)
				if err != nil {
					return fmt.Errorf("could not decode req body json: %w", err)
				}
				inputVal = reflect.ValueOf(inputValPtr)
			} else {
				mForm, err := c.MultipartForm()
				if err != nil {
					return fmt.Errorf("cannot get multipart form: %w", err)
				}
				err = g.processMultipartUpload(mForm, inputVal, fParams)
				if err != nil {
					return fmt.Errorf("cannot process multipart form: %w", err)
				}
			}

		}
		inputVal = inputVal.Elem()
		for _, pParamName := range pathParams {
			paramVal := c.Param(pParamName.ParamName)
			fItem := inputVal.FieldByName(pParamName.StructFieldName)
			fItem.Set(reflect.ValueOf(paramVal))
		}
		for _, qParamName := range queryParams {
			paramVal := c.QueryParam(qParamName.ParamName)
			fItem := inputVal.FieldByName(qParamName.StructFieldName)
			fItem.Set(reflect.ValueOf(paramVal))
		}

		inValues := make([]reflect.Value, 0)
		inValues = append(inValues, reflect.ValueOf(c.Request().Context()))
		inValues = append(inValues, inputVal)
		if inParamsCount > 2 {
			inValues = append(inValues, reflect.ValueOf(c.Request()))
		}
		if inParamsCount > 3 {
			inValues = append(inValues, reflect.ValueOf(c.Response()))
		}
		results := handlerFunc.Call(inValues)
		var errVal reflect.Value
		if outParamsCount == 1 {
			errVal = results[0]
		} else {
			errVal = results[1]
		}
		errInterface := errVal.Interface()
		var resultErr error = nil
		if errInterface != nil {
			var ok bool
			resultErr, ok = errInterface.(error)
			if !ok {
				return fmt.Errorf("calling %s: callback cannot process error: %v", handlerType.String(), errInterface)
			}
		}

		if resultErr != nil {
			return c.JSON(http.StatusInternalServerError, resultErr)
		}

		if outParamsCount == 2 {
			output := results[0].Interface()
			return c.JSON(http.StatusOK, output)
		}
		return nil
	}
}

func (g *WrapGroup) getUploadFileParams(paramType reflect.Type) []fileField {
	fParams := make([]fileField, 0)
	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		if field.Type == fHeaderType {
			fInfo := generator.GetFieldInfo(field)
			if fInfo == nil {
				continue
			}
			fP := fileField{
				fieldName:       fInfo.Name,
				structFieldName: field.Name,
				multiple:        false,
			}
			fParams = append(fParams, fP)
			continue
		} else if field.Type.Kind() == reflect.Slice {
			sElem := field.Type.Elem()
			if sElem == fHeaderType {
				fInfo := generator.GetFieldInfo(field)
				if fInfo == nil {
					continue
				}
				fP := fileField{
					fieldName:       fInfo.Name,
					structFieldName: field.Name,
					multiple:        true,
				}
				fParams = append(fParams, fP)
			}
		}
	}
	return fParams
}

func (g *WrapGroup) processMultipartUpload(mForm *multipart.Form, inputVal reflect.Value, fParams []fileField) error {
	reqBody := mForm.Value["request"]
	if len(reqBody) > 0 {
		reqVal := []byte(reqBody[0])
		inputValPtr := inputVal.Interface()
		err := json.Unmarshal(reqVal, inputValPtr)
		if err != nil {
			return fmt.Errorf("could not decode req body json: %w", err)
		}
		inputVal = reflect.ValueOf(inputValPtr)
	}
	inputVal = inputVal.Elem()
	for _, fP := range fParams {
		fHeaders, ok := mForm.File[fP.fieldName]
		if !ok {
			continue
		}
		fItem := inputVal.FieldByName(fP.structFieldName)
		if fP.multiple {
			fItem.Set(reflect.ValueOf(fHeaders))
		} else {
			if len(fHeaders) > 0 {
				fItem.Set(reflect.ValueOf(fHeaders[0]))
			}
		}

	}
	return nil
}

func (g *WrapGroup) getParams(paramType reflect.Type, pathParams []string, processBody bool) (qParamNames []ReqField, pParamNames []ReqField) {
	for i := 0; i < paramType.NumField(); i++ {
		field := paramType.Field(i)
		fInfo := generator.GetFieldInfo(field)
		if fInfo == nil {
			continue
		}
		if field.Type.Kind() != reflect.String {
			// Only type string supported for query parameters
			continue
		}
		var isPathParam bool
		if fInfo.In == "path" {
			isPathParam = true
		} else {
			for _, pParam := range pathParams {
				if fInfo.Name == pParam {
					isPathParam = true
				}
			}
		}
		reqField := ReqField{
			ParamName:       fInfo.Name,
			StructFieldName: field.Name,
		}
		if isPathParam {
			pParamNames = append(pParamNames, reqField)
		} else if fInfo.In == "query" || !processBody {
			qParamNames = append(qParamNames, reqField)
		}
	}
	return qParamNames, pParamNames
}
