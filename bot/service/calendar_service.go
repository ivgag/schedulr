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

func (s *CalendarService) CreateEvent(userID int, event domain.Event) error {
	linkedAccount, err := s.userService.GetGoogleAccount(userID)
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
		s.userService.UpdateGoogleAccessToken(userID, *refreshedToken)
	}

	return err
}
