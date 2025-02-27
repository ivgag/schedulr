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

func NewLinkedAccountRepository(db *sql.DB) LinkedAccountRepository {
	return &PgLinkedAccountRepository{db: db}
}

type PgLinkedAccountRepository struct {
	db *sql.DB
}

// GetByUserID implements LinkedAccountRepository.
func (p *PgLinkedAccountRepository) GetByUserID(userID int) ([]model.LinkedAccount, error) {
	rows, err := p.db.Query(`
	SELECT id, user_id, provider, access_token, refresh_token, expiry
	FROM linked_accounts
	WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.LinkedAccount
	for rows.Next() {
		var account model.LinkedAccount
		if err := rows.Scan(&account.ID, &account.UserID, &account.Provider, &account.AccessToken, &account.RefreshToken, &account.Expiry); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Save implements ConnectedAccountRepository.
func (p *PgLinkedAccountRepository) Save(account *model.LinkedAccount) error {
	row := p.db.QueryRow(`
	INSERT INTO linked_accounts(user_id, provider, access_token, refresh_token, expiry, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5, timezone('utc', now()), timezone('utc', now()))
	ON CONFLICT (user_id, provider) DO UPDATE
	SET access_token = EXCLUDED.access_token,
		refresh_token = EXCLUDED.refresh_token,
		expiry = EXCLUDED.expiry,
		updated_at = timezone('utc', now())
	RETURNING id
	`,
		account.UserID, account.Provider, account.AccessToken, account.RefreshToken, account.Expiry,
	)

	if err := row.Scan(&account.ID); err != nil {
		return err
	}

	return nil
}

// GetByUserIDAndProvider implements ConnectedAccountRepository.
func (p *PgLinkedAccountRepository) GetByUserIDAndProvider(userID int, provider model.Provider) (model.LinkedAccount, error) {
	var account model.LinkedAccount

	err := p.db.QueryRow(`
	SELECT id, user_id, provider, access_token, refresh_token, expiry 
	FROM linked_accounts
	WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	).Scan(&account.ID, &account.UserID, &account.Provider, &account.AccessToken, &account.RefreshToken, &account.Expiry)

	if err != nil && err.Error() == noRowsError {
		return model.LinkedAccount{}, model.NotFoundError{Message: "account not found"}
	} else if err != nil {
		return model.LinkedAccount{}, err
	} else {
		return account, nil
	}
}

func (p *PgLinkedAccountRepository) DeleteByUserIDAndProvider(userID int, provider model.Provider) (bool, error) {
	result, err := p.db.Exec(`
	DELETE FROM linked_accounts
	WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)

	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}
