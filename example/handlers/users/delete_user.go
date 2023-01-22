package users

import (
	"context"
	"example_http_server/store"
	"github.com/AlhimicMan/goswag/wrapper"
	"net/http"
)

type DeleteUserReq struct {
	ID string `json:"id" validate:"uuid"`
}

type DeleteUserRes struct {
}

func DeleteUser(ctx context.Context, req DeleteUserReq, httpReq *http.Request) (DeleteUserRes, *wrapper.ErrorResult) {
	err := store.DeleteUser(req.ID)
	if err != nil {
		errRes := wrapper.ErrorResult{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
		return DeleteUserRes{}, &errRes
	}
	return DeleteUserRes{}, nil
}
