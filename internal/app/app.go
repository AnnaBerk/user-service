package app

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"user-service/config"
	v1 "user-service/controller/v1"
	"user-service/pkg/httpserver"
	"user-service/pkg/kafka"
	"user-service/pkg/psql"
	"user-service/pkg/redis"
	"user-service/repo/pgdb"
	"user-service/service"
)

type Person struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}

func Run() {
	// конфигурации
	cfg := config.LoadConfig()

	// Logger
	SetLogrus(cfg.Log.Level)

	// Создаем подключение к  PostgreSQL
	storage, err := psql.New(cfg.ConnectionString, psql.MaxPoolSize(cfg.MaxPoolSize))
	if err != nil {
		log.Fatal("failed to init storage", err)
		os.Exit(1)
	}
	defer storage.Close()

	// Инициализируем Kafka
	kafkaService := kafka.New(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer func() {
		if err := kafkaService.Close(); err != nil {
			log.Fatalf("Failed to close Kafka service: %s", err)
		}
	}()

	// Инициализируем редис
	redisClient := redis.New(cfg.Redis.Addr, cfg.Redis.Password)

	// создаем экземпляр сервиса с зависимостями
	userRepo := pgdb.NewUserRepo(storage)
	fioService := service.NewFIOService(kafkaService, userRepo, redisClient)

	// запускаем основной цикл обработки сообщений
	go fioService.ProcessMessages()

	// Echo
	log.Info("Initializing handlers and routes...")
	handler := echo.New()

	// эндпоинты
	v1.NewRouter(handler, fioService)

	// HTTP сервер
	log.Info("Starting http server...")
	httpServer := httpserver.New(handler, httpserver.Port(cfg.Port))

	log.Info("Configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error("app - Run - httpServer.Notify: %w", err)
	}

	log.Info("Shutting down...")
	err = httpServer.Shutdown()
	if err != nil {
		log.Error("app - Run - httpServer.Shutdown: %w", err)
	}

}
