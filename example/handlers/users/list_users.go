package users

import (
	"context"
	"example_http_server/store"
	"net/http"
)

type ListUsersReq struct {
	Offset int
	Limit  int
}

type OneListedUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

type ListUserRes struct {
	Users []OneListedUser `json:"users"`
}

func ListUsers(ctx context.Context, req ListUsersReq, httpReq *http.Request) (ListUserRes, error) {
	listed := store.ListUses()
	resUsers := make([]OneListedUser, 0, len(listed))
	for _, usr := range listed {
		if req.Offset > 0 {
			req.Offset -= 1
			continue
		}
		resRec := OneListedUser{
			Login: usr.Login,
			Name:  usr.Name,
		}
		resUsers = append(resUsers, resRec)
		if len(resUsers) == req.Limit {
			break
		}
	}
	return ListUserRes{Users: resUsers}, nil
}
