package storage

import "time"

type ConnectedAccount struct {
	ID             int
	UserID         int
	Provider       string
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt time.Time
}

type ConnectedAccountRepository interface {
	CreateConnectedAccount(account *ConnectedAccount) error
	GetConnectedAccountByUserIDAndProvider(userID int, provider string) (ConnectedAccount, error)
}
