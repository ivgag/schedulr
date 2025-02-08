package service

import (
	"github.com/ivgag/schedulr/storage"
)

func NewUserService(userRepository storage.UserRepository) *UserService {
	return &UserService{}
}

type UserService struct {
	userRepository storage.UserRepository
}

func (s *UserService) GetUserByID(id int) (storage.User, error) {
	return s.userRepository.GetUserByID(id)
}

func (s *UserService) CreateUser(user *storage.User) error {
	return s.userRepository.CreateUser(user)
}
