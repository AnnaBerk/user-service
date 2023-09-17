package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"user-service/api_clients/model"
	"user-service/pkg/kafka"
	"user-service/repo"
)

type FIOService struct {
	kafkaService *kafka.Service
	userRepo     repo.UserRepo
	RedisClient  *redis.Client
	stopCh       chan bool
}

type FIO struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

// устанавливаем срок жизни кэша
const CacheExpiration = 5 * time.Minute

func NewFIOService(kafkaService *kafka.Service, userRepo repo.UserRepo, rdb *redis.Client) *FIOService {
	return &FIOService{
		kafkaService: kafkaService,
		userRepo:     userRepo,
		stopCh:       make(chan bool),
		RedisClient:  rdb,
	}
}

// ProcessMessages читает очередь кафки, валидирует сообщение и в случае успеха записывает результат в бд
func (f *FIOService) ProcessMessages() {
	message := FIO{
		Name:       "Petr",
		Surname:    "Ushakov",
		Patronymic: "Vasilevich",
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Error("Failed to marshal the message:", err)
		return
	}

	if err := f.kafkaService.PublishToTopic("FIO", messageBytes); err != nil {
		log.Error("Failed to publish the message to Kafka:", err)
	}
	for {
		select {
		case <-f.stopCh:
			// Если мы получаем сигнал остановки, завершаем функцию
			return
		default:
			msg, err := f.kafkaService.ReadMessage(context.Background())
			if err != nil {
				// Обрабатываем ошибку чтения из Kafka
				log.Error("Failed to read message from Kafka:", err)
				continue
			}

			// Десериализация сообщения
			var fioMessage FIO
			err = json.Unmarshal(msg.Value, &fioMessage)
			if err != nil {
				// Отправляем сообщение в очередь FIO_FAILED
				if pubErr := f.kafkaService.PublishToTopic("FIO_FAILED", msg.Value); pubErr != nil {
					log.Error("Failed to publish message to FIO_FAILED queue:", pubErr)
				}
				continue
			}

			// Валидация сообщения
			err = fioMessage.IsValid()
			if err != nil {
				errorMsg := map[string]interface{}{
					"error":            err.Error(),
					"original_message": fioMessage,
				}
				errorBytes, marshalErr := json.Marshal(errorMsg)
				if marshalErr != nil {
					log.Error("Failed to marshal error message:", marshalErr)
					continue
				}
				// Отправляем сообщение об ошибке в очередь FIO_FAILED
				if pubErr := f.kafkaService.PublishToTopic("FIO_FAILED", errorBytes); pubErr != nil {
					log.Error("Failed to publish message to FIO_FAILED queue:", pubErr)
				}
				continue
			}

			// Обогащение информации
			enrichedData, err := f.enrichFIOData(fioMessage)
			if err != nil {
				log.Error("Failed to enrich the FIO data:", err)
				continue
			}

			// Сохранение в БД
			ctx := context.Background()
			user := convertToUser(enrichedData)
			err = f.userRepo.Save(ctx, user)
			if err != nil {
				log.Error("Failed to save user to the database:", err)
			}
		}
	}
}

// GetUsers получает пользователей по заданным параметрам с пагинацией
// также тут реализован пример использования кэша
func (f *FIOService) GetUsers(c echo.Context) error {
	pageStr := c.QueryParam("page")
	sizeStr := c.QueryParam("size")
	filter := c.QueryParam("filter")

	// Преобразование из строки в int
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid page parameter")
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid size parameter")
	}

	// Составляем ключ для кеширования
	cacheKey := fmt.Sprintf("users:p=%d:s=%d:f=%s", page, size, filter)
	cachedData, err := f.RedisClient.Get(c.Request().Context(), cacheKey).Result()

	if err == nil {
		// Если в кеше есть данные, десериализуем их и возвращаем
		var users []model.User
		err = json.Unmarshal([]byte(cachedData), &users)
		if err == nil {
			return c.JSON(http.StatusOK, users)
		}
	}

	users, err := f.userRepo.GetUsers(page, size, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch users",
		})
	}

	// Сериализуем новые данные и сохраняем их в кеш
	data, _ := json.Marshal(users)
	f.RedisClient.Set(c.Request().Context(), cacheKey, data, CacheExpiration)

	return c.JSON(http.StatusOK, users)
}

// AddUser добавляет пользователя, обязательные параметры name и surname, возвращает объект добавленного пользователя
//
//go:generate mockgen -destination=./mocks/user_repo_mock.go -package=mocks user-service/repo UserRepo
func (f *FIOService) AddUser(c echo.Context) error {
	user := model.User{}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse request body",
		})
	}
	// Проверка на наличие name и username
	if user.Name == "" || user.Surname == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Both name and username are required",
		})
	}

	id, err := f.userRepo.AddUser(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add user",
		})
	}
	user.ID = id
	return c.JSON(http.StatusCreated, user)
}

// DeleteUser удаляет пользователя по заданному id
func (f *FIOService) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid size parameter")
	}

	err = f.userRepo.DeleteUser(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete user",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

// UpdateUser обновляет пользователя переданными параметрам по id, если пользователя с id нет, возвращает ошибку
func (f *FIOService) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid ID parameter")
	}

	user := model.User{}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse request body",
		})
	}

	user.ID = id

	err = f.userRepo.UpdateUser(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update user",
		})
	}
	return c.JSON(http.StatusOK, user)
}

func convertToUser(data EnrichedFIO) model.User {
	// Простой пример: взять страну с наибольшей вероятностью.
	// На практике можно добавить дополнительную логику или обработку ошибок.
	var country string
	if len(data.Nationality) > 0 {
		country = data.Nationality[0].CountryID
	}

	return model.User{
		Name:        data.Name,
		Surname:     data.Surname,
		Patronymic:  data.Patronymic,
		Age:         data.Age,
		Gender:      data.Gender,
		Nationality: country,
	}
}
