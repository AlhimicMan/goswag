package generator

import "reflect"

const (
	BasicAuth = "basic"
	OAuth2    = "oauth2"
	APIKey    = "apiKey"
)

type HandlerInfo struct {
	RequestType *reflect.Type
	OutputType  *reflect.Type
	FileUpload  []FileUploadParameters
}

type FileUploadParameters struct {
	Name          string
	jsonName      string
	MultipleFiles bool
}

type BasicAuthParams struct {
}

type OAuth2Params struct {
	Flow             string
	AuthorizationURL string
	TokenURL         string
	Scopes           map[string]string
}

type APIKeyParams struct {
	In   string
	Name string
}

type AuthType struct {
	AuthTypeName string
	Description  string
	Scopes       []string
	BasicAuth    *BasicAuthParams
	OAuth2       *OAuth2Params
	APIKey       *APIKeyParams
}

type HandlerParameters struct {
	Summary    string
	Auth       []AuthType
	FileUpload []FileUploadParameters
}

type RouteInfo struct {
	Method     string
	Tags       []string
	Handler    HandlerInfo
	Parameters HandlerParameters
}

type fieldInfo struct {
	Name     string
	JSONName string
	In       string
}
