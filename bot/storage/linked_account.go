package storage

import "time"

type LinkedAccount struct {
	ID           int
	UserID       int
	Provider     string
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type LinkedAccountRepository interface {
	Create(account *LinkedAccount) error
	Update(account *LinkedAccount) error
	GetByUserIDAndProvider(userID int, provider string) (LinkedAccount, error)
}
