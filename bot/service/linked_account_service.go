package service

import (
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/storage"
)

func NewLinkedAccountService() *LinkedAccountService {
	return &LinkedAccountService{}
}

type LinkedAccountService struct {
	linkedAccountRepository storage.LinkedAccountRepository
}

func (s *LinkedAccountService) GetLinkedAccountsByUserID(userID int) ([]model.LinkedAccount, error) {
	return s.linkedAccountRepository.GetByUserID(userID)
}

func (s *LinkedAccountService) GetLinkedAccountByProviderAndUserID(
	userID int,
	provider model.Provider,
) (model.LinkedAccount, error) {
	return s.linkedAccountRepository.GetByUserIDAndProvider(userID, provider)
}

func (s *LinkedAccountService) SaveLinkedAccount(account *model.LinkedAccount) error {
	return s.linkedAccountRepository.Save(account)
}

func (s *LinkedAccountService) DeleteLinkedAccount(userID int, provider model.Provider) (bool, error) {
	return s.linkedAccountRepository.DeleteByUserIDAndProvider(userID, provider)
}
