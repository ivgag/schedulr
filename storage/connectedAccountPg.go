package storage

import (
	"database/sql"
)

func NewPostgresConnectedAccountRepository(db *sql.DB) ConnectedAccountRepository {
	return &PostgresConnectedAccountRepository{db: db}
}

type PostgresConnectedAccountRepository struct {
	db *sql.DB
}

// CreateConnectedAccount implements ConnectedAccountRepository.
func (p *PostgresConnectedAccountRepository) CreateConnectedAccount(account *ConnectedAccount) error {
	row := p.db.QueryRow(`
	INSERT INTO connected_accounts(user_id, provider, access_token, refresh_token, token_expires_at) 
	VALUES($1, $2, $3, $4, $5) 
	RETURNING id`,
		account.UserID, account.Provider, account.AccessToken, account.RefreshToken, account.TokenExpiresAt,
	)

	if err := row.Scan(&account.ID); err != nil {
		return err
	}

	return nil
}

// GetConnectedAccountByUserIDAndProvider implements ConnectedAccountRepository.
func (p *PostgresConnectedAccountRepository) GetConnectedAccountByUserIDAndProvider(userID int, provider string) (ConnectedAccount, error) {
	row := p.db.QueryRow(`
	SELECT id, user_id, provider, access_token, refresh_token, token_expires_at 
	FROM connected_accounts 
	WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)

	return ConnectedAccount{}, row.Err()
}
