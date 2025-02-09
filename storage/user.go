// storage/user.go
package storage

type User struct {
	ID         int
	TelegramId int64
}

type UserRepository interface {
	GetUserByID(id int) (User, error)
	CreateUser(user *User) error
}
