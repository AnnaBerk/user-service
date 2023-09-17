package service

import (
	"github.com/labstack/echo/v4"
)

type FIOServiceInterface interface {
	ProcessMessages()
	GetUsers(c echo.Context) error
	AddUser(c echo.Context) error
	DeleteUser(c echo.Context) error
	UpdateUser(c echo.Context) error
}
