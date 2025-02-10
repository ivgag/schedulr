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
	connectedAccountRepository storage.ConnectedAccountRepository,
) *UserService {
	return &UserService{
		gClient:                    gClient,
		userRepository:             userRepository,
		connectedAccountRepository: connectedAccountRepository,
	}
}

type UserService struct {
	gClient                    google.GoogleClient
	userRepository             storage.UserRepository
	connectedAccountRepository storage.ConnectedAccountRepository
}

func (s *UserService) GetUserByID(id int) (storage.User, error) {
	return s.userRepository.GetUserByID(id)
}

func (s *UserService) GetUserByTelegramID(telegramID int64) (storage.User, error) {
	return s.userRepository.GetUserByTelegramID(telegramID)
}

func (s *UserService) CreateUser(user *storage.User) error {
	return s.userRepository.CreateUser(user)
}

func (s *UserService) GetGoogleConnectionUrl(userID int) string {
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

	err = s.connectedAccountRepository.CreateConnectedAccount(&storage.ConnectedAccount{
		UserID:         userID,
		Provider:       "google",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenExpiresAt: token.Expiry,
	})

	return err
}
