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
	query := "SELECT id, telegram_id FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.TelegramId)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// GetUserByTelegramID implements UserRepository.
func (r *PostgresUserRepository) GetUserByTelegramID(telegramID int64) (User, error) {
	var user User
	query := "SELECT id, telegram_id FROM users WHERE telegram_id = $1"
	err := r.db.QueryRow(query, telegramID).Scan(&user.ID, &user.TelegramId)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *PostgresUserRepository) CreateUser(user *User) error {
	query := "INSERT INTO users(telegram_id) VALUES($1) RETURNING id"
	return r.db.QueryRow(query, user.TelegramId).Scan(&user.ID)
}
