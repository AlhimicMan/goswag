package users

import (
	"context"
	"example_http_server/store"
	"net/http"
)

type GetUserAvatarReq struct {
	ID string `json:"id" validate:"uuid"`
}

func GetUserAvatar(ctx context.Context, req GetUserAvatarReq, _ *http.Request, respWriter http.ResponseWriter) error {
	avatar, err := store.GetAvatarByUserID(req.ID)
	if err != nil {
		return err
	}
	respWriter.Write(avatar.Data)
	return nil
}
