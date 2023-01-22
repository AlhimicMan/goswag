package users

import (
	"context"
	"example_http_server/store"
)

type UpdateUserReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func UpdateUser(ctx context.Context, req UpdateUserReq) (UserRec, error) {
	usr, err := store.GetUserByID(req.ID)
	if err != nil {
		return UserRec{}, err
	}
	usr.Name = req.Name
	err = store.UpdateUser(usr)
	if err != nil {
		return UserRec{}, err
	}
	rec := UserRec{
		Login: usr.Login,
		Name:  usr.Name,
		Email: usr.Email,
	}
	return rec, nil
}
