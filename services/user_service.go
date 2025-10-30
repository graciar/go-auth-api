package services

import (
	"context"
	"go-auth/models"
)

type UserService interface {
	Signup(context.Context, *models.User) error
	EmailExists(context.Context, string) (bool, error)
	Login(context.Context, *string, *string) (string, string, *models.User, error)

	SaveOTP(context.Context, string, string) error
	VerifyOTP(context.Context, string, string) error
	ResetPassword(context.Context, string, string) error

	Refresh(context.Context, string) (string, string, error)

	GetUser(context.Context, *string) (*models.User, error)
	GetAll(context.Context, int, int, int) ([]*models.User, error)

	UpdateUser(context.Context, *models.User) error
	DeleteUser(context.Context, string) error
}
