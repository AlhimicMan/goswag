package users

import (
	"context"
	"example_http_server/store"
	"fmt"
	"net/http"
	"strconv"
)

type GetUserReq struct {
	ID     string `json:"id" validate:"uuid"`
	Public string `param:"public,query"`
}

type UserRec struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetUser(ctx context.Context, req GetUserReq, httpReq *http.Request) (UserRec, error) {
	user, err := store.GetUserByID(req.ID)
	if err != nil {
		return UserRec{}, err
	}
	res := UserRec{
		Login: user.Login,
		Name:  user.Name,
	}
	pub, err := strconv.ParseBool(req.Public)
	if err != nil {
		return UserRec{}, fmt.Errorf("invalid value for parameter public")
	}
	if !pub {
		res.Email = user.Email
	}
	return res, nil
}
