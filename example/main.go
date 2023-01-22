package main

import (
	"example_http_server/handlers/users"
	"github.com/AlhimicMan/goswag/generator"
	"github.com/AlhimicMan/goswag/wrapper"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	router := wrapper.NewRouter(e)
	group := router.Group("/users", "Users")
	RegisterRoutes(group)
	for _, route := range e.Routes() {
		e.Logger.Infof("registered %s: %s %s", route.Method, route.Path, route.Name)
	}
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	_, err := router.GenerateSwagger()
	if err != nil {
		e.Logger.Fatalf("cannot generate swagger: %w", err)
	}

	e.Logger.Fatal(e.Start(":1323"))
}

func RegisterRoutes(group *wrapper.WrapGroup) {
	authParams := generator.AuthType{
		AuthTypeName: "API Key",
		Description:  "API key authentication",
		APIKey:       &generator.APIKeyParams{In: "header", Name: "X-API-Key"},
	}
	hAuth := []generator.AuthType{authParams}
	group.GET("/get/:id", generator.HandlerParameters{
		Summary: "Get user",
		Auth:    hAuth,
	}, users.GetUser)
	group.GET("/get/:id/avatar", generator.HandlerParameters{
		Summary: "Get user avatar",
		Auth:    hAuth,
	}, users.GetUserAvatar)
	group.GET("/list", generator.HandlerParameters{
		Summary: "List users",
	}, users.ListUsers)
	additionalFilesUploadParams := generator.FileUploadParameters{Name: "custom_file", MultipleFiles: false}
	group.POST("/create", generator.HandlerParameters{
		Summary:    "Create user",
		Auth:       hAuth,
		FileUpload: []generator.FileUploadParameters{additionalFilesUploadParams},
	}, users.CreateUser)
	group.POST("/update/:id", generator.HandlerParameters{
		Summary: "Update user",
		Auth:    hAuth,
	}, users.UpdateUser)
	group.POST("/update/:id/avatar", generator.HandlerParameters{
		Summary: "Update user avatar",
		Auth:    hAuth,
	}, users.UpdateUserAvatar)
	group.DELETE("/delete/:id", generator.HandlerParameters{
		Summary: "Delete user",
		Auth:    hAuth,
	}, users.DeleteUser)
}
