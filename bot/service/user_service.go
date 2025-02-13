package service

import (
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/storage"
)

func NewUserService(
	userRepository storage.UserRepository,
	tokenServices map[model.Provider]TokenService,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		tokenServices:  tokenServices,
	}
}

type UserService struct {
	userRepository storage.UserRepository
	tokenServices  map[model.Provider]TokenService
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

func (s *UserService) GetOAuth2Url(telegramID int64, provider model.Provider) (string, error) {
	user, err := s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return "", err
	} else if user.ID == 0 {
		return "", errors.New("user not found")
	}

	return s.tokenServices[provider].GetOAuth2URL(user.ID)
}

func (s *UserService) LinkAccount(state string, provider model.Provider, code string) error {
	return s.tokenServices[provider].ExchangeCodeForToken(state, code)
}
