package wrapper

import (
	"fmt"
	"github.com/AlhimicMan/goswag/generator"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func (g *WrapGroup) Group(prefix string, tag string, m ...echo.MiddlewareFunc) *WrapGroup {
	group := &WrapGroup{
		echoGroup:      g.echoGroup.Group(prefix, m...),
		path:           g.path + prefix,
		routesHandlers: make(map[string]generator.RouteInfo),
		tags:           []string{tag},
	}
	g.childGroups[prefix] = group
	return group
}

func (g *WrapGroup) POST(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	// Here we get definitions from handler and add them to routes
	handlerInfo, err := processHandler(handler)
	fullPath := g.path + path
	if err == nil {
		routeInfo := generator.RouteInfo{
			Method:     http.MethodPost,
			Handler:    handlerInfo,
			Tags:       g.tags,
			Parameters: params,
		}
		handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
		g.routesHandlers[handlerKey] = routeInfo
	}
	echoHandler := g.callProcessor(fullPath, handler, true)
	return g.echoGroup.POST(path, echoHandler, m...)
}

func (g *WrapGroup) GET(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	handlerInfo, err := processHandler(handler)
	fullPath := g.path + path
	if err == nil {
		routeInfo := generator.RouteInfo{
			Method:     http.MethodGet,
			Handler:    handlerInfo,
			Tags:       g.tags,
			Parameters: params,
		}
		handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
		g.routesHandlers[handlerKey] = routeInfo
	}
	echoHandler := g.callProcessor(fullPath, handler, false)
	return g.echoGroup.GET(path, echoHandler, m...)
}

func (g *WrapGroup) DELETE(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	handlerInfo, err := processHandler(handler)
	fullPath := g.path + path
	if err == nil {
		routeInfo := generator.RouteInfo{
			Method:     http.MethodDelete,
			Handler:    handlerInfo,
			Tags:       g.tags,
			Parameters: params,
		}
		handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
		g.routesHandlers[handlerKey] = routeInfo
	}
	echoHandler := g.callProcessor(fullPath, handler, false)
	return g.echoGroup.DELETE(path, echoHandler, m...)
}

func (g *WrapGroup) CONNECT(path string, _ generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	echoHandler := g.callProcessor(path, handler, false)
	return g.echoGroup.CONNECT(path, echoHandler, m...)
}

func (g *WrapGroup) HEAD(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	handlerInfo := generator.HandlerInfo{}
	fullPath := g.path + path
	routeInfo := generator.RouteInfo{
		Method:     http.MethodHead,
		Handler:    handlerInfo,
		Tags:       g.tags,
		Parameters: params,
	}
	handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
	g.routesHandlers[handlerKey] = routeInfo
	echoHandler := g.callProcessor(fullPath, handler, false)
	return g.echoGroup.HEAD(path, echoHandler, m...)
}

func (g *WrapGroup) OPTIONS(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	handlerInfo := generator.HandlerInfo{}
	fullPath := g.path + path
	routeInfo := generator.RouteInfo{
		Method:     http.MethodOptions,
		Handler:    handlerInfo,
		Tags:       g.tags,
		Parameters: params,
	}
	handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
	g.routesHandlers[handlerKey] = routeInfo
	echoHandler := g.callProcessor(path, handler, false)
	return g.echoGroup.OPTIONS(path, echoHandler, m...)
}

func (g *WrapGroup) PATCH(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	handlerInfo, err := processHandler(handler)
	fullPath := g.path + path
	if err == nil {
		routeInfo := generator.RouteInfo{
			Method:     http.MethodPatch,
			Handler:    handlerInfo,
			Tags:       g.tags,
			Parameters: params,
		}
		handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
		g.routesHandlers[handlerKey] = routeInfo
	}
	echoHandler := g.callProcessor(fullPath, handler, true)
	return g.echoGroup.PATCH(path, echoHandler, m...)
}

func (g *WrapGroup) PUT(path string, params generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	handlerInfo, err := processHandler(handler)
	fullPath := g.path + path
	if err == nil {
		routeInfo := generator.RouteInfo{
			Method:     http.MethodPut,
			Handler:    handlerInfo,
			Tags:       g.tags,
			Parameters: params,
		}
		handlerKey := fmt.Sprintf("%s~%s", routeInfo.Method, fullPath)
		g.routesHandlers[handlerKey] = routeInfo
	}
	echoHandler := g.callProcessor(fullPath, handler, true)
	return g.echoGroup.PATCH(path, echoHandler, m...)
}

func (g *WrapGroup) TRACE(path string, _ generator.HandlerParameters, handler interface{}, m ...echo.MiddlewareFunc) *echo.Route {
	echoHandler := g.callProcessor(path, handler, false)
	return g.echoGroup.TRACE(path, echoHandler, m...)
}

func (g *WrapGroup) getRoutes() map[string]generator.RouteInfo {
	routes := make(map[string]generator.RouteInfo)
	for _, group := range g.childGroups {
		groupRoutes := group.getRoutes()
		for path, routeInfo := range groupRoutes {
			routes[path] = routeInfo
		}
	}
	for path, routeInfo := range g.routesHandlers {
		routes[path] = routeInfo
	}
	return routes
}
