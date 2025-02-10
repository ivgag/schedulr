package storage

import (
	"database/sql"
)

func NewLinkedAccountRepository(db *sql.DB) LinkedAccountRepository {
	return &PgLinkedAccountRepository{db: db}
}

type PgLinkedAccountRepository struct {
	db *sql.DB
}

// Create implements ConnectedAccountRepository.
func (p *PgLinkedAccountRepository) Create(account *LinkedAccount) error {
	row := p.db.QueryRow(`
	INSERT INTO linked_accounts(user_id, provider, access_token, refresh_token, token_expires_at) 
	VALUES($1, $2, $3, $4, $5) 
	RETURNING id`,
		account.UserID, account.Provider, account.AccessToken, account.RefreshToken, account.TokenExpiresAt,
	)

	if err := row.Scan(&account.ID); err != nil {
		return err
	}

	return nil
}

// GetByUserIDAndProvider implements ConnectedAccountRepository.
func (p *PgLinkedAccountRepository) GetByUserIDAndProvider(userID int, provider string) (LinkedAccount, error) {
	row := p.db.QueryRow(`
	SELECT id, user_id, provider, access_token, refresh_token, token_expires_at 
	FROM linked_accounts
	WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)

	return LinkedAccount{}, row.Err()
}
