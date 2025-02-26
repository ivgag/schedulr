package model

import "time"

type UserProfile struct {
	User           User
	LinkedAccounts []LinkedAccount
}

func (p UserProfile) LinkedAccount(provider Provider) *LinkedAccount {
	for _, account := range p.LinkedAccounts {
		if account.Provider == provider {
			return &account
		}
	}
	return nil
}

type User struct {
	ID         int
	TelegramID int64
	Username   string
	Timezone   string
}

type LinkedAccount struct {
	ID           int
	UserID       int
	AccountName  string
	Provider     Provider
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}
