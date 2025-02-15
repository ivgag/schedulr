package storage

import (
	"time"

	"github.com/ivgag/schedulr/model"
)

type LinkedAccount struct {
	ID           int
	UserID       int
	Provider     model.Provider
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type LinkedAccountRepository interface {
	Save(account LinkedAccount) error
	GetByUserIDAndProvider(userID int, provider model.Provider) (LinkedAccount, error)
}
