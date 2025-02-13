package storage

import (
	"time"

	"github.com/ivgag/schedulr/domain"
)

type LinkedAccount struct {
	ID           int
	UserID       int
	Provider     domain.Provider
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type LinkedAccountRepository interface {
	Create(account *LinkedAccount) error
	Update(account *LinkedAccount) error
	GetByUserIDAndProvider(userID int, provider domain.Provider) (LinkedAccount, error)
}
