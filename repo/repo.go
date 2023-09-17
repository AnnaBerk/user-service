package repo

import (
	"context"
	"user-service/api_clients/model"
)

type UserRepo interface {
	Save(ctx context.Context, user model.User) error
	GetUsers(page, size int, filter string) ([]model.User, error)
	AddUser(user model.User) (int, error)
	DeleteUser(id int) error
	UpdateUser(user model.User) error
}
