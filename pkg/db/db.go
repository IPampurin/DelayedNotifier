package db

import (
	"time"

	"github.com/IPampurin/DelayedNotifier/pkg/configuration"
)

type Notification struct {
	ID          int       // идентификатор уведомления, назначается сервисом в момент вызова POST /notify
	User_id     int       // адресат уведомления (кому)
	Channel     []string  // канал доставки (email/telegram)
	Content     string    // само уведомление
	Status      string    // текущий статус (scheduled/sent/failed)
	Send_for    time.Time // когда планируется отправить (гггг.мм.дд чч:мм:сс)
	Send_at     time.Time // фактическое время отправки (гггг.мм.дд чч:мм:сс), момент получения Consumer подтверждения от внешнего API
	Retry_count int       // счетчик попыток отправки (для Consumer)
	Last_error  string    // информация о сбое при крайней отправке
	Created_at  time.Time // время создания
}

func InitDB(cfgDB *configuration.ConfDB) error {

	return nil
}

/*
// CloseDB закрывает соединение с базой
func CloseDB() {

	// получаем объект *sql.DB для закрытия соединения
	sqlDB, err := DB.Db.DB()
	if err != nil {
		log.Printf("Ошибка при получении SQL соединения: %v", err)
		return
	}

	// закрываем соединение
	if err := sqlDB.Close(); err != nil {
		log.Printf("Предупреждение: ошибка при закрытии БД: %v", err)
	} else {
		log.Println("БД успешно отключена.")
	}
}
*/
