package model

import (
	"maragu.dev/glue/model"
)

type AccountID = model.AccountID

type Account struct {
	ID      AccountID
	Created Time
	Updated Time
	Name    string
}

type UserID = model.UserID

type User struct {
	ID        UserID
	Created   Time
	Updated   Time
	AccountID AccountID `db:"account_id"`
	Name      string
	Email     EmailAddress
	Confirmed bool
	Active    bool
}

type Role = model.Role

const (
	RoleAdmin Role = "admin"
)

type Permission = model.Permission
