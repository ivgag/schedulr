// storage/user.go
package storage

type User struct {
	ID         int
	TelegramId int64
}

type UserRepository interface {
	GetByID(id int) (User, error)
	GetByTelegramID(telegramID int64) (User, error)
	Create(user *User) error
}
