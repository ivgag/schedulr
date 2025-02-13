package service

import (
	"errors"

	"github.com/gofrs/uuid"
	"github.com/ivgag/schedulr/domain"
	"github.com/ivgag/schedulr/google"
	"github.com/ivgag/schedulr/storage"
)

var stateTokens = make(map[string]int)
var stateProvider = make(map[string]domain.Provider)

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

func (s *UserService) GetOAuth2Url(telegramID int64, provider domain.Provider) (string, error) {
	user, err := s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return "", err
	} else if user.ID == 0 {
		return "", errors.New("user not found")
	}

	googleAccount, err := s.linkedAccountRepository.GetByUserIDAndProvider(user.ID, provider)
	if err != nil {
		return "", err
	} else if googleAccount.ID != 0 {
		return "", errors.New(string(provider) + " account already linked")
	}

	state, err := uuid.NewGen().NewV7()
	if err != nil {
		return "", err
	}

	stateTokens[state.String()] = user.ID
	stateProvider[state.String()] = provider

	return s.gClient.GetOAuth2Url(state.String()), nil
}

func (s *UserService) LinkAccount(state string, code string) error {
	userID, ok := stateTokens[state]
	if !ok {
		return errors.New("invalid state token")
	}

	token, err := s.gClient.ExchangeCodeForToken(code)
	if err != nil {
		return err
	}

	err = s.linkedAccountRepository.Create(&storage.LinkedAccount{
		UserID:       userID,
		Provider:     stateProvider[state],
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})

	delete(stateTokens, state)
	delete(stateProvider, state)

	return err
}

func (s *UserService) GetLinkedAccount(userID int, provider domain.Provider) (storage.LinkedAccount, error) {
	return s.linkedAccountRepository.GetByUserIDAndProvider(userID, provider)
}

func (s *UserService) UpdateLinkedAccountAccessToken(
	userID int,
	provider domain.Provider,
	token domain.Token,
) error {
	account, err := s.linkedAccountRepository.GetByUserIDAndProvider(userID, provider)
	if err != nil {
		return err
	}

	account.AccessToken = token.AccessToken
	account.RefreshToken = token.RefreshToken
	account.Expiry = token.Expiry

	return s.linkedAccountRepository.Update(&account)
}
