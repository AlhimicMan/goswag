package wrapper

import (
	"encoding/json"
	"github.com/AlhimicMan/goswag/generator"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/swaggo/swag"
)

type RouteWrapper struct {
	router *echo.Echo
	groups []*WrapGroup
}

func NewRouter(router *echo.Echo) *RouteWrapper {
	return &RouteWrapper{
		router: router,
		groups: make([]*WrapGroup, 0),
	}
}

func (s *RouteWrapper) Group(prefix string, tag string, m ...echo.MiddlewareFunc) (g *WrapGroup) {
	group := &WrapGroup{
		echoGroup:      s.router.Group(prefix, m...),
		path:           prefix,
		routesHandlers: make(map[string]generator.RouteInfo),
		tags:           []string{tag},
	}
	s.groups = append(s.groups, group)
	return group
}

func (s *RouteWrapper) getRoutes() map[string]generator.RouteInfo {
	routes := make(map[string]generator.RouteInfo)
	for _, group := range s.groups {
		groupRoutes := group.getRoutes()
		for path, routeInfo := range groupRoutes {
			routes[path] = routeInfo
		}
	}
	return routes
}

func (s *RouteWrapper) GenerateSwagger() ([]byte, error) {
	routes := s.getRoutes()
	gen := generator.NewSwaggerGenerator()
	swagSpec, err := gen.EmitOpenAPIDefinition(routes)
	if err != nil {
		return nil, err
	}
	swagSpec.Info.Title = "Portal API"
	jsonBytes, err := json.MarshalIndent(swagSpec, "", "    ")
	if err != nil {
		return nil, errors.Errorf("Cannot marshal Swagger annotation: %w", err)
	}
	s.registerDefinition(string(jsonBytes))
	return jsonBytes, nil
}

func (s *RouteWrapper) registerDefinition(template string) {
	var SwaggerInfo = &swag.Spec{
		InfoInstanceName: "swagger",
		SwaggerTemplate:  template,
	}
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)

}
