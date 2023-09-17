package service

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-service/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
)

func TestAddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepo(ctrl)
	e := echo.New()
	f := &FIOService{userRepo: mockUserRepo}

	t.Run("Failed to parse request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("invalid_json")))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = f.AddUser(c) // Вызываем метод и игнорируем ошибку, так как она будет возвращена в HTTP ответе

		if got, want := rec.Code, http.StatusBadRequest; got != want {
			t.Errorf("got status %d, wanted %d", got, want)
		}
	})

	t.Run("Both name and surrname are required", func(t *testing.T) {
		userJSON := `{"Name": "", "Surname": ""}`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(userJSON)))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = f.AddUser(c)

		if got, want := rec.Code, http.StatusBadRequest; got != want {
			t.Errorf("got status %d, wanted %d", got, want)
		}
	})

	t.Run("Failed to add user", func(t *testing.T) {
		mockUserRepo.EXPECT().AddUser(gomock.Any()).Return(0, errors.New("DB error"))

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"Name": "John", "Surname": "Smith"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = f.AddUser(c)

		if got, want := rec.Code, http.StatusInternalServerError; got != want {
			t.Errorf("got status %d, wanted %d", got, want)
		}
	})

	t.Run("User added successfully", func(t *testing.T) {
		mockUserRepo.EXPECT().AddUser(gomock.Any()).Return(1, nil)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(`{"Name": "Franz", "Surname": "Kafka"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = f.AddUser(c)

		if got, want := rec.Code, http.StatusCreated; got != want {
			t.Errorf("got status %d, wanted %d", got, want)
		}
	})
}
