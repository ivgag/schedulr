package storage

import (
	"database/sql"

	"github.com/ivgag/schedulr/model"
)

type PgUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) GetByID(id int) (User, error) {
	var user User
	query := "SELECT id, telegram_id FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.TelegramID)

	if err != nil && err.Error() == noRowsError {
		return User{}, model.NotFoundError{Message: "user not found"}
	} else if err != nil {
		return User{}, err
	} else {
		return user, nil
	}
}

// GetByTelegramID implements UserRepository.
func (r *PgUserRepository) GetByTelegramID(telegramID int64) (User, error) {
	var user User
	query := "SELECT id, telegram_id FROM users WHERE telegram_id = $1"
	err := r.db.QueryRow(query, telegramID).Scan(&user.ID, &user.TelegramID)

	if err != nil && err.Error() == noRowsError {
		return User{}, model.NotFoundError{Message: "user not found"}
	} else if err != nil {
		return User{}, err
	} else {
		return user, nil
	}
}

func (r *PgUserRepository) Save(user *User) error {
	query := `
	INSERT INTO users(telegram_id)
	VALUES($1)
	ON CONFLICT (telegram_id)
	DO UPDATE SET telegram_id = users.telegram_id
	RETURNING id;
	`
	return r.db.QueryRow(query, user.TelegramID).Scan(&user.ID)
}
