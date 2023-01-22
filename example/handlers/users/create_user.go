package users

import (
	"context"
	"example_http_server/models"
	"example_http_server/store"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type CreateUserReq struct {
	Login    string                  `json:"login" validate:"required"`
	Password string                  `json:"password" validate:"required"`
	Name     string                  `json:"name" validate:"required"`
	Email    string                  `json:"email"  validate:"email"`
	Avatar   *multipart.FileHeader   `param:"avatar"`
	Docs     []*multipart.FileHeader `param:"docs"`
}

type CreateUserRes struct {
	ID              string `json:"id"`
	DocsCount       int    `json:"docs_count"`
	AdditionalFound bool   `json:"additional_found"`
}

func CreateUser(ctx context.Context, req CreateUserReq, httpReq *http.Request) (CreateUserRes, error) {
	userRec := models.UserRec{
		Login:    req.Login,
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}
	err := store.CreateUser(&userRec)
	if err != nil {
		return CreateUserRes{}, err
	}
	if req.Avatar != nil {
		avatarFile, err := req.Avatar.Open()
		if err != nil {
			return CreateUserRes{}, fmt.Errorf("cannot open avatar file: %w", err)
		}
		fileData, err := io.ReadAll(avatarFile)
		if err != nil {
			return CreateUserRes{}, fmt.Errorf("canot read avatar file: %w", err)
		}
		avatarRec := models.Avatar{Data: fileData}
		err = store.SaveAvatarByUserID(userRec.ID, avatarRec)
		if err != nil {
			return CreateUserRes{}, fmt.Errorf("cannot dave uploaded avatar %w", err)
		}
	}
	var additionalFound bool
	_, _, err = httpReq.FormFile("custom_file")
	if err == nil {
		additionalFound = true
	}
	res := CreateUserRes{
		ID:              userRec.ID,
		DocsCount:       len(req.Docs),
		AdditionalFound: additionalFound,
	}
	return res, nil
}
