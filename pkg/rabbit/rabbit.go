package rabbit

import (
	"context"
	"fmt"
	"time"

	"github.com/IPampurin/DelayedNotifier/pkg/configuration"
	"github.com/wb-go/wbf/logger"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

// ClientRabbit хранит ссылку на оригинальный клиент RabbitMQ и имя очереди.
type ClientRabbit struct {
	*rabbitmq.RabbitClient
	queue string
}

// глобальный экземпляр клиента (синглтон)
var defaultClient *ClientRabbit

// InitRabbit инициализирует подключение к RabbitMQ, декларирует очередь и сохраняет глобальный клиент
func InitRabbit(cfg *configuration.ConfRabbitMQ, consumerCfg *configuration.ConfConsumer, log logger.Logger) error {

	// формируем URL для подключения
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%d%s", cfg.User, cfg.Password, cfg.HostName, cfg.Port, cfg.VHost)

	// создаём стратегии повторных попыток
	reconnectStrategy := retry.Strategy{
		Attempts: consumerCfg.RetryCount,
		Delay:    consumerCfg.RetryDelay,
		Backoff:  float64(consumerCfg.Backoff),
	}
	// для публикации и потребления можно использовать ту же стратегию.
	producingStrategy := reconnectStrategy
	consumingStrategy := reconnectStrategy

	// конфигурация клиента RabbitMQ
	clientCfg := rabbitmq.ClientConfig{
		URL:            amqpURL,
		ConnectionName: "delayed-notifier",
		ConnectTimeout: 10 * time.Second,
		Heartbeat:      30 * time.Second,
		ReconnectStrat: reconnectStrategy, // стратегия переподключения при обрыве
		ProducingStrat: producingStrategy, // стратегия повторов при публикации
		ConsumingStrat: consumingStrategy, // стратегия повторов при обработке (для консумера)
	}

	// создаём клиента
	client, err := rabbitmq.NewClient(clientCfg)
	if err != nil {
		return fmt.Errorf("ошибка создания клиента RabbitMQ: %w", err)
	}

	// декларируем очередь, будем использовать прямую очередь без exchange
	// передаём exchangeName = "" и routingKey = queueName,
	// очередь делаем долговременной (durable = true), неавтоудаляемой
	if err := client.DeclareQueue(cfg.Queue, "", cfg.Queue, true, false, true, nil); err != nil {
		// закрываем клиент при ошибке
		_ = client.Close()
		return fmt.Errorf("ошибка создания очереди RabbitMQ: %w", err)
	}

	defaultClient = &ClientRabbit{
		RabbitClient: client,
		queue:        cfg.Queue,
	}

	log.Info("RabbitMQ запущен", "queue", cfg.Queue)
	return nil
}

// GetClient возвращает глобальный экземпляр клиента Rabbit
func GetClient() *ClientRabbit {
	return defaultClient
}

// Publish публикует сообщение в заданную очередь (Publisher с exchange = "")
func (c *ClientRabbit) Publish(ctx context.Context, body []byte) error {
	publisher := rabbitmq.NewPublisher(c.RabbitClient, "", "application/json")
	// публикуем напрямую в очередь, routingKey = имя очереди
	return publisher.Publish(ctx, body, c.queue)
}

// GetQueueName возвращает имя очереди (будет для консумера)
func (c *ClientRabbit) GetQueueName() string {
	return c.queue
}
