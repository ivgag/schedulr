package service

import (
	"github.com/ivgag/schedulr/domain"
	"github.com/ivgag/schedulr/google"
)

func NewCalendarService(
	gClient google.GoogleClient,
	userService UserService,
) *CalendarService {
	return &CalendarService{
		gClient:     gClient,
		userService: userService,
	}
}

type CalendarService struct {
	gClient     google.GoogleClient
	userService UserService
}

func (s *CalendarService) CreateEvent(userID int, provider domain.Provider, event domain.Event) error {
	linkedAccount, err := s.userService.GetLinkedAccount(userID, provider)
	if err != nil {
		return err
	}

	token := domain.Token{
		AccessToken:  linkedAccount.AccessToken,
		RefreshToken: linkedAccount.RefreshToken,
		Expiry:       linkedAccount.Expiry,
	}

	refreshedToken, err := s.gClient.CreateEvent(token, event)

	if err != nil {
		return err
	}

	if token != *refreshedToken {
		s.userService.UpdateLinkedAccountAccessToken(userID, provider, *refreshedToken)
	}

	return err
}
