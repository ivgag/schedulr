// storage/user.go
package storage

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	GetUserByID(id int) (User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
}
