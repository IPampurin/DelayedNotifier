package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/IPampurin/DelayedNotifier/pkg/configuration"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
)

// Notification — структура, описывающая уведомление и соответствующая строке в таблице notifications БД
type Notification struct {
	UID        uuid.UUID    // uid, UUID для глобальной уникальности
	UserID     int          // user_id, адресат уведомления (кому)
	Channel    []string     // channel, канал доставки (email/telegram)
	Content    string       // content, само уведомление
	Status     string       // status, текущий статус (scheduled/sent/failed)
	SendFor    time.Time    // send_for, когда планируется отправить (гггг.мм.дд чч:мм:сс)
	SendAt     sql.NullTime // send_at, фактическое время отправки (гггг.мм.дд чч:мм:сс), момент получения Consumer подтверждения от внешнего API
	RetryCount int          // retry_count, счетчик попыток отправки (для Consumer)
	LastError  string       // last_error, информация о сбое при крайней отправке
	CreatedAt  time.Time    // created_at, время создания
}

// Client хранит подключение к БД
// делаем его публичным, чтобы другие пакеты могли использовать методы
type Client struct {
	*dbpg.DB
}

// глобальный экземпляр клиента (синглтон)
var defaultClient *Client

// InitDB инициализирует подключение к PostgreSQL
func InitDB(cfg *configuration.ConfDB) error {

	// формируем DSN для master (и slaves, если потребуется)
	// формат: postgres://user:pass@host:port/dbname?sslmode=disable
	masterDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.HostName, cfg.Port, cfg.Name)

	// опции подключения (можно вынести и в конфиг)
	opts := &dbpg.Options{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	// создаём подключение (пока без слейвов)
	dbConn, err := dbpg.New(masterDSN, nil, opts)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	defaultClient = &Client{dbConn}

	// применяем миграции
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := defaultClient.Migration(ctx); err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	return nil
}

// CloseDB закрывает соединение с БД
func CloseDB() error {

	if defaultClient != nil && defaultClient.Master != nil {
		return defaultClient.Master.Close()
	}

	return nil
}

// GetClient возвращает глобальный экземпляр клиента БД
func GetClient() *Client {

	return defaultClient
}

// CreateNotification создаёт новое уведомление в БД
func (c *Client) CreateNotification(ctx context.Context, n *Notification) error {

	query := `INSERT INTO notifications (uid, user_id, channel, content, status, send_for, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := c.ExecContext(ctx, query, n.UID, n.UserID, dbpg.Array(&n.Channel), n.Content, n.Status, n.SendFor, n.CreatedAt)

	return err
}

// GetNotification возвращает уведомление по UID
func (c *Client) GetNotification(ctx context.Context, uid uuid.UUID) (*Notification, error) {

	query := `SELECT uid, user_id, channel,
	                 content, status, send_for, send_at,
					 retry_count, last_error, created_at
	            FROM notifications
			   WHERE uid = $1`
	var n Notification
	var channelSlice []string
	err := c.QueryRowContext(ctx, query, uid).Scan(
		&n.UID, &n.UserID, dbpg.Array(&channelSlice),
		&n.Content, &n.Status, &n.SendFor, &n.SendAt,
		&n.RetryCount, &n.LastError, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	n.Channel = channelSlice

	return &n, nil
}

// CancelNotification помечает уведомление как отменённое (статус cancelled)
// возвращает sql.ErrNoRows, если уведомление с таким uid не найдено или уже не в статусе 'scheduled'
func (c *Client) CancelNotification(ctx context.Context, uid uuid.UUID) error {

	query := `UPDATE notifications
	             SET status = 'cancelled'
			   WHERE uid = $1 AND status = 'scheduled'`
	result, err := c.ExecContext(ctx, query, uid)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows // ошибка "уведомление не найдено или уже не scheduled"
	}

	return nil
}

// UpdateNotificationStatus обновляет статус и связанные поля (для consumer)
// sentAt может быть nil, если отправка ещё не производилась
func (c *Client) UpdateNotificationStatus(ctx context.Context, uid uuid.UUID, status string, sentAt *time.Time, retryCount int, lastError string) error {

	query := `UPDATE notifications
                 SET status = $1, send_at = $2, retry_count = $3, last_error = $4
		       WHERE uid = $5`
	_, err := c.ExecContext(ctx, query, status, sentAt, retryCount, lastError, uid)

	return err
}
