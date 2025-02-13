package model

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expires_at"`
	Provider     Provider  `json:"provider"`
}

type Provider string

const (
	ProviderGoogle Provider = "google"
)
