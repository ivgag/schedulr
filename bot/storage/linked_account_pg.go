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

// Create implements ConnectedAccountRepository.
func (p *PgLinkedAccountRepository) Create(account LinkedAccount) error {
	row := p.db.QueryRow(`
	INSERT INTO linked_accounts(user_id, provider, access_token, refresh_token, expiry) 
	VALUES($1, $2, $3, $4, $5) 
	RETURNING id`,
		account.UserID, account.Provider, account.AccessToken, account.RefreshToken, account.Expiry,
	)

	if err := row.Scan(&account.ID); err != nil {
		return err
	}

	return nil
}

// Update implements LinkedAccountRepository.
func (p *PgLinkedAccountRepository) Update(account LinkedAccount) error {
	_, err := p.db.Exec(`
	UPDATE linked_accounts
	SET access_token = $1, refresh_token = $2, expiry = $3
	WHERE id = $4`,
		account.AccessToken, account.RefreshToken, account.Expiry, account.ID,
	)

	return err
}

// GetByUserIDAndProvider implements ConnectedAccountRepository.
func (p *PgLinkedAccountRepository) GetByUserIDAndProvider(userID int, provider model.Provider) (LinkedAccount, error) {
	var account LinkedAccount

	err := p.db.QueryRow(`
	SELECT id, user_id, provider, access_token, refresh_token, expiry 
	FROM linked_accounts
	WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	).Scan(&account.ID, &account.UserID, &account.Provider, &account.AccessToken, &account.RefreshToken, &account.Expiry)

	if err != nil && err.Error() == noRowsError {
		return LinkedAccount{}, model.NotFoundError{Message: "account not found"}
	} else if err != nil {
		return LinkedAccount{}, err
	} else {
		return account, nil
	}
}
