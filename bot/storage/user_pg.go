/*
 * Created on Mon Feb 17 2025
 *
 *  Copyright (c) 2025 Ivan Gagarkin
 * SPDX-License-Identifier: EPL-2.0
 *
 * Licensed under the Eclipse Public License - v 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.eclipse.org/legal/epl-2.0/
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func (r *PgUserRepository) GetByID(id int) (model.User, error) {
	var user model.User
	query := "SELECT id, telegram_id, username, timezone, preferred_calendar FROM users WHERE id = $1"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.TelegramID, &user.Username)

	if err != nil && err.Error() == noRowsError {
		return model.User{}, model.NotFoundError{Message: "user not found"}
	} else if err != nil {
		return model.User{}, err
	} else {
		return user, nil
	}
}

// GetByTelegramID implements UserRepository.
func (r *PgUserRepository) GetByTelegramID(telegramID int64) (model.User, error) {
	var user model.User
	query := "SELECT id, telegram_id, username, timezone, preferred_calendar FROM users WHERE telegram_id = $1"
	err := r.db.QueryRow(
		query,
		telegramID,
	).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.TimeZone,
		&user.PreferredCalendar,
	)

	if err != nil && err.Error() == noRowsError {
		return model.User{}, model.NotFoundError{Message: "user not found"}
	} else if err != nil {
		return model.User{}, err
	} else {
		return user, nil
	}
}

func (r *PgUserRepository) Save(user *model.User) error {
	query := `
	INSERT INTO users(telegram_id, username, timezone, preferred_calendar)
	VALUES($1, $2, $3, $4)
	ON CONFLICT (telegram_id)
	DO UPDATE SET telegram_id = users.telegram_id, 
		username = EXCLUDED.username, 
		timezone = EXCLUDED.timezone,
		preferred_calendar = EXCLUDED.preferred_calendar
	RETURNING id;
	`
	return r.db.QueryRow(
		query,
		user.TelegramID,
		user.Username,
		user.TimeZone,
		user.PreferredCalendar,
	).Scan(&user.ID)
}
