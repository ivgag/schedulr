package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/gofrs/uuid"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/storage"
	"github.com/ivgag/schedulr/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var stateTokens = make(map[string]int)

// NewGoogleTokenService creates a new TokenService.
func NewGoogleTokenService(
	linkedAccountsRepository storage.LinkedAccountRepository,
) (*GoogleTokenService, error) {
	clientID, err := utils.GetenvOrError("GOOGLE_CLIENT_ID")
	if err != nil {
		return nil, err
	}

	clientSecret, err := utils.GetenvOrError("GOOGLE_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}

	redirectURL, err := utils.GetenvOrError("GOOGLE_REDIRECT_URL")
	if err != nil {
		return nil, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{calendar.CalendarScope},
	}

	return &GoogleTokenService{
		oauth2Config:             oauth2Config,
		linkedAccountsRepository: linkedAccountsRepository,
	}, nil
}

// GoogleTokenService encapsulates OAuth2 token logic.
type GoogleTokenService struct {
	oauth2Config             *oauth2.Config
	linkedAccountsRepository storage.LinkedAccountRepository
}

// GetOAuth2URL returns the URL to redirect users for Google OAuth2 consent.
func (s *GoogleTokenService) GetOAuth2URL(userID int) (string, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	state := uuid.String()
	stateTokens[state] = userID

	return s.oauth2Config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	), nil
}

// ExchangeCodeForToken exchanges an authorization code for a token.
func (s *GoogleTokenService) ExchangeCodeForToken(state string, code string) error {
	usedID, found := stateTokens[state]
	if !found {
		return fmt.Errorf("state not found: %s", state)
	}

	gToken, err := s.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	err = s.linkedAccountsRepository.Save(storage.LinkedAccount{
		UserID:       usedID,
		Provider:     model.ProviderGoogle,
		AccessToken:  gToken.AccessToken,
		RefreshToken: gToken.RefreshToken,
		Expiry:       gToken.Expiry.UTC(),
	})

	return err
}

// ClientFromToken creates an HTTP client authenticated with the given token.
func (s *GoogleTokenService) ClientForUser(userID int) (*http.Client, error) {
	account, err := s.linkedAccountsRepository.GetByUserIDAndProvider(userID, model.ProviderGoogle)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(account.Expiry) {
		tokenSource := s.oauth2Config.TokenSource(context.Background(), &oauth2.Token{
			RefreshToken: account.RefreshToken,
		})
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, err
		}

		account.AccessToken = newToken.AccessToken
		account.RefreshToken = newToken.RefreshToken
		account.Expiry = newToken.Expiry

		err = s.linkedAccountsRepository.Save(account)
		if err != nil {
			return nil, err
		}
	}

	oauthToken := &oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		Expiry:       account.Expiry,
		TokenType:    "Bearer",
	}
	return s.oauth2Config.Client(context.Background(), oauthToken), nil
}

// CalendarService handles calendar-related operations.
type GoogleCalendarService struct {
	tokenService *GoogleTokenService
}

// NewGoogleCalendarService creates a new GoogleCalendarService.
func NewGoogleCalendarService(tokenService *GoogleTokenService) *GoogleCalendarService {
	return &GoogleCalendarService{tokenService: tokenService}
}

// CreateEvent creates a new calendar event using the provided token and event data.
func (c *GoogleCalendarService) CreateEvent(userID int, event *model.Event) (*model.Event, error) {
	client, err := c.tokenService.ClientForUser(userID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	cal, err := srv.Calendars.Get("primary").Do()
	if err != nil {
		return nil, err
	}

	calEvent := toGoogleCalendarEvent(event)

	link, err := c.insertEventWithRetries(srv, calEvent)
	if err != nil {
		log.Error().
			Interface("event", calEvent).
			Err(err).
			Msg("Failed to create event")

		return nil, err
	}

	event.Link = link
	event.Start.TimeZone = cal.TimeZone
	event.End.TimeZone = cal.TimeZone

	return event, nil
}

func (c *GoogleCalendarService) insertEventWithRetries(
	srv *calendar.Service,
	event *calendar.Event,
) (string, error) {
	operation := func() (string, error) {
		createdEvent, err := srv.Events.Insert("primary", event).Do()
		if err != nil {
			return "", err
		}

		return createdEvent.HtmlLink, nil
	}

	return backoff.Retry(
		context.Background(),
		operation,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(3),
	)
}

func toGoogleCalendarEvent(event *model.Event) *calendar.Event {
	return &calendar.Event{
		Summary:     event.Title,
		Location:    event.Location,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.Start.DateTime.Format(time.RFC3339),
			TimeZone: event.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: event.End.DateTime.Format(time.RFC3339),
			TimeZone: event.End.TimeZone,
		},
	}
}
