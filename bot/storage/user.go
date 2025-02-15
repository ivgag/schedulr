package storage

type User struct {
	ID         int
	TelegramID int64
}

type UserRepository interface {
	GetByID(id int) (User, error)
	GetByTelegramID(telegramID int64) (User, error)
	Save(user *User) error
}
