// storage/postgres_user.go
package storage

import (
	"database/sql"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) GetUserByID(id int) (User, error) {
	var user User
	query := "SELECT id, name FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *PostgresUserRepository) CreateUser(user *User) error {
	query := "INSERT INTO users(name) VALUES($1) RETURNING id"
	return r.db.QueryRow(query, user.Name).Scan(&user.ID)
}

func (r *PostgresUserRepository) UpdateUser(user *User) error {
	query := "UPDATE users SET name = $1 WHERE id = $2"
	return r.db.QueryRow(query, user.Name).Scan(&user.ID)
}
