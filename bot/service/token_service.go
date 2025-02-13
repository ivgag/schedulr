package service

type TokenService interface {
	GetOAuth2URL(userID int) (string, error)
	ExchangeCodeForToken(state string, code string) error
}
