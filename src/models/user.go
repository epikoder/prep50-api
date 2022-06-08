package models

import "github.com/google/uuid"

type (
	IUser interface {
		ID() uuid.UUID
		Name() string
		UserName() string
		GetEmail() string
	}

	User struct {
		Id    uuid.UUID
		FName string
		LName string
		UName string
		Email string
	}
)

func (u *User) ID() (id uuid.UUID) {
	return
}
func (u *User) Name() (name string) {
	return
}
func (u *User) UserName() (username string) {
	return
}
func (u *User) GetEmail() (email string) {
	return
}
