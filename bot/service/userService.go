package service

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/ivgag/schedulr/google"
	"github.com/ivgag/schedulr/storage"
)

var stateTokens = make(map[string]int)

func NewUserService(
	gClient google.GoogleClient,
	userRepository storage.UserRepository,
	connectedAccountRepository storage.LinkedAccountRepository,
) *UserService {
	return &UserService{
		gClient:                 gClient,
		userRepository:          userRepository,
		linkedAccountRepository: connectedAccountRepository,
	}
}

type UserService struct {
	gClient                 google.GoogleClient
	userRepository          storage.UserRepository
	linkedAccountRepository storage.LinkedAccountRepository
}

func (s *UserService) GetUserByID(id int) (storage.User, error) {
	return s.userRepository.GetByID(id)
}

func (s *UserService) GetUserByTelegramID(telegramID int64) (storage.User, error) {
	return s.userRepository.GetByTelegramID(telegramID)
}

func (s *UserService) CreateUser(user *storage.User) error {
	return s.userRepository.Create(user)
}

func (s *UserService) GetOAuth2Url(userID int) string {
	state, err := uuid.NewGen().NewV7()
	if err != nil {
		panic(err)
	}

	stateTokens[state.String()] = userID

	return s.gClient.GetCalendarConnectionUrl(state.String())
}

func (s *UserService) ConnectGoogleAccount(state string, code string) error {
	userID, ok := stateTokens[state]
	if !ok {
		return errors.New("invalid state token")
	}

	delete(stateTokens, state)

	token, err := s.gClient.ExchangeCodeForToken(code)
	if err != nil {
		return err
	}

	err = s.linkedAccountRepository.Create(&storage.LinkedAccount{
		UserID:         userID,
		Provider:       "google",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenExpiresAt: token.Expiry,
	})

	return err
}
