package store

import (
	"errors"
	"example_http_server/models"
	"github.com/google/uuid"
)

var usersStorage map[string]models.UserRec
var loginToID map[string]string
var avatarStorage map[string]models.Avatar

func init() {
	usersStorage = make(map[string]models.UserRec)
	avatarStorage = make(map[string]models.Avatar)
	loginToID = make(map[string]string)
}

func CreateUser(user *models.UserRec) error {
	_, ok := loginToID[user.Login]
	if ok {
		return errors.New("login duplicated")
	}
	id := uuid.New().String()
	user.ID = id
	usersStorage[id] = *user
	loginToID[user.Login] = user.ID
	return nil
}

func GetUserByID(id string) (models.UserRec, error) {
	user, ok := usersStorage[id]
	if !ok {
		return models.UserRec{}, errors.New("user not found")
	}
	return user, nil
}

func UpdateUser(user models.UserRec) error {
	_, ok := usersStorage[user.ID]
	if !ok {
		return errors.New("user not found")
	}
	usersStorage[user.ID] = user
	return nil
}

func ListUses() []models.UserRec {
	res := make([]models.UserRec, 0, len(usersStorage))
	for _, usr := range usersStorage {
		res = append(res, usr)
	}
	return res
}

func DeleteUser(id string) error {
	usr, ok := usersStorage[id]
	if !ok {
		return errors.New("user not found")
	}
	delete(usersStorage, id)
	delete(loginToID, usr.Login)
	return nil
}

func SaveAvatarByUserID(id string, avatar models.Avatar) error {
	avatarStorage[id] = avatar
	return nil
}

func GetAvatarByUserID(id string) (models.Avatar, error) {
	avatar, ok := avatarStorage[id]
	if !ok {
		return models.Avatar{}, errors.New("avatar for user not defined")
	}
	return avatar, nil
}
