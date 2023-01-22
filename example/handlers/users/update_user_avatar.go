package users

import (
	"context"
	"example_http_server/models"
	"example_http_server/store"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type UpdateUserAvatarReq struct {
	ID     string                `json:"id"`
	Force  string                `param:"force,query" json:"force"`
	Avatar *multipart.FileHeader `json:"avatar"`
}

type UpdateUserAvatarRes struct {
	Forced bool `json:"forced"`
}

func UpdateUserAvatar(ctx context.Context, req UpdateUserAvatarReq, httpReq *http.Request) (UpdateUserAvatarRes, error) {
	force, err := strconv.ParseBool(req.Force)
	if err != nil {
		return UpdateUserAvatarRes{}, fmt.Errorf("invalid value for parameter force")
	}
	avatarFile, err := req.Avatar.Open()
	if err != nil {
		return UpdateUserAvatarRes{}, fmt.Errorf("cannot open avatar file: %w", err)
	}
	fileData, err := io.ReadAll(avatarFile)
	if err != nil {
		return UpdateUserAvatarRes{}, fmt.Errorf("canot read avatar file: %w", err)
	}
	avatarRec := models.Avatar{Data: fileData}
	err = store.SaveAvatarByUserID(req.ID, avatarRec)
	if err != nil {
		return UpdateUserAvatarRes{}, fmt.Errorf("cannot dave uploaded avatar %w", err)
	}
	res := UpdateUserAvatarRes{
		Forced: force,
	}
	return res, nil
}
