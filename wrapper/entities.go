package wrapper

import (
	"github.com/AlhimicMan/goswag/generator"
	"github.com/labstack/echo/v4"
)

type WrapGroup struct {
	echoGroup *echo.Group

	path           string
	routesHandlers map[string]generator.RouteInfo
	tags           []string
	childGroups    map[string]*WrapGroup
}

type EmptyReq struct{}
type EmptyResp struct{}
