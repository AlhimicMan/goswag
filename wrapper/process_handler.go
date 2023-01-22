package wrapper

import (
	"context"
	"github.com/AlhimicMan/goswag/generator"
	"github.com/pkg/errors"
	"net/http"
	"reflect"
)

func processHandler(handler interface{}) (generator.HandlerInfo, error) {
	handlerType := reflect.TypeOf(handler)

	inputParamsCount := handlerType.NumIn()
	if inputParamsCount < 2 || inputParamsCount > 4 {
		return generator.HandlerInfo{}, errors.Errorf("cannot register handler: unsupported input params count: %d", inputParamsCount)
	}
	// Check pattern ctx, params, request, responseWriter
	ctxParam := handlerType.In(0)
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !ctxParam.Implements(ctxType) {
		return generator.HandlerInfo{}, errors.Errorf("cannot register handler: first parameter must be Context")
	}

	reqParam := handlerType.In(1)
	if reqParam.Kind() != reflect.Struct {
		return generator.HandlerInfo{}, errors.Errorf("cannot register handler: second parameter must be struct")
	}

	if inputParamsCount > 2 {
		requestType := reflect.TypeOf(http.Request{})
		requestParamType := handlerType.In(2).Elem()
		if requestType != requestParamType {
			return generator.HandlerInfo{}, errors.Errorf("cannot register handler: third parameter must be *http.Request, have %s", requestParamType.String())
		}
	}

	if inputParamsCount > 3 {
		respWriterType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
		respParam := handlerType.In(3)
		if !respParam.Implements(respWriterType) {
			return generator.HandlerInfo{}, errors.Errorf("cannot register handler: fourth parameter must implement http.ResponseWriter")
		}
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	outParamsCount := handlerType.NumOut()
	var outType *reflect.Type
	if outParamsCount == 1 {
		if !handlerType.Out(0).Implements(errorInterface) {
			panic("Second return value should be an error")
		}
	} else if outParamsCount == 2 {
		if handlerType.Out(0).Kind() != reflect.Struct {
			panic("First return value be a struct")
		}
		if !handlerType.Out(1).Implements(errorInterface) {
			panic("Second return value should be an error")
		}
		outRes := handlerType.Out(0)
		outType = &outRes
	} else {
		return generator.HandlerInfo{}, errors.Errorf("cannot register handler: unsupported out params count %d", outParamsCount)
	}

	handlerInfo := generator.HandlerInfo{
		RequestType: &reqParam,
		OutputType:  outType,
	}
	return handlerInfo, nil
}
