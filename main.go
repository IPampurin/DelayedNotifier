package main

import (
	"log"
	"os"

	"github.com/IPampurin/DelayedNotifier/pkg/configuration"
	"github.com/IPampurin/DelayedNotifier/pkg/db"
	"github.com/IPampurin/DelayedNotifier/pkg/server"
	"github.com/wb-go/wbf/logger"
)

func main() {

	var err error

	// считываем .env файл
	cfg, err := configuration.ReadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// настраиваем логгер
	appLogger, err := logger.InitLogger(
		logger.ZapEngine,
		"delayed-notifier",
		os.Getenv("APP_ENV"), // пока оставим пустым
		logger.WithLevel(logger.InfoLevel),
	)
	if err != nil {
		log.Fatalf("Ошибка создания логгера: %v", err)
	}
	defer func() { _ = appLogger.(*logger.ZapAdapter) }()

	// подключаем базу данных
	err = db.InitDB(&cfg.DB)
	if err != nil {
		appLogger.Error("ошибка подключения к БД", "error", err)
		return
	}
	defer db.CloseDB()

	// инициализируем кэш
	err = cache.InitCache(&cfg.Redis)
	if err != nil {
		appLogger.Warn("кэш не работает", "error", err)
	}

	// запускаем RabbitMQ

	// запускаем консумер

	// запускаем сервер
	if err := server.Run(&cfg.Server, appLogger); err != nil {
		appLogger.Error("Ошибка сервера", "error", err)
		return
	}

	appLogger.Info("Приложение корректно завершено")
}
