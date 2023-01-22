package models

type UserRec struct {
	ID       string
	Login    string
	Email    string
	Name     string
	Password string
}

type Avatar struct {
	Data []byte
}
