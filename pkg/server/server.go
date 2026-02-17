package server

import (
	"fmt"
	"time"

	"github.com/IPampurin/DelayedNotifier/pkg/api"
	"github.com/IPampurin/DelayedNotifier/pkg/configuration"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

func Run(cfgServer *configuration.ConfServer, log logger.Logger) error {

	// создаём движок Gin через обёртку ginext
	engine := ginext.New(cfgServer.GinMode)

	// добавляем стандартные middleware (логгер и восстановление)
	engine.Use(ginext.Logger(), ginext.Recovery())

	// добавляем свой middleware для структурного логирования запросов
	engine.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		// используем переданный логгер для записи информации о запросе
		log.LogRequest(c.Request.Context(), c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	})

	// регистрируем эндпоинты согласно заданию

	apiGroup := engine.Group("/notify")
	{
		apiGroup.POST("/", api.CreateNotificationHandler(log))
		apiGroup.GET("/:id", api.GetNotificationHandler(log))
		apiGroup.DELETE("/:id", api.DeleteNotificationHandler(log))
	}

	// раздаём статические файлы из папки ./web
	engine.Static("/static", "./web")

	// для корневого пути отдаём index.html
	engine.GET("/", func(c *ginext.Context) {
		c.File("./web/index.html")
	})

	// формируем адрес запуска
	addr := fmt.Sprintf("%s:%d", cfgServer.HostName, cfgServer.Port)
	log.Info("запуск HTTP-сервера", "address", addr)

	// запускаем сервер (блокирующий вызов)
	return engine.Run(addr)
}
