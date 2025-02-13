package google

import (
	"context"
	"fmt"

	"github.com/ivgag/schedulr/domain"
	"github.com/ivgag/schedulr/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func NewGoogleClient() (GoogleClient, error) {
	clientID, err := utils.GetenvOrError("GOOGLE_CLIENT_ID")
	if err != nil {
		return GoogleClient{}, err
	}

	clientSecret, err := utils.GetenvOrError("GOOGLE_CLIENT_SECRET")
	if err != nil {
		return GoogleClient{}, err
	}

	redirectURL, err := utils.GetenvOrError("GOOGLE_REDIRECT_URL")
	if err != nil {
		return GoogleClient{}, err
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{calendar.CalendarScope},
	}

	return GoogleClient{
		oauth2Config: oauth2Config,
	}, nil
}

type GoogleClient struct {
	oauth2Config *oauth2.Config
}

func (c *GoogleClient) GetOAuth2Url(oauthStateString string) string {
	return c.oauth2Config.AuthCodeURL(
		oauthStateString,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"), // Always ask for refresh token.
	)
}

func (c *GoogleClient) ExchangeCodeForToken(code string) (*domain.Token, error) {
	gToken, err := c.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  gToken.AccessToken,
		RefreshToken: gToken.RefreshToken,
		Expiry:       gToken.Expiry,
	}, nil
}

func (c *GoogleClient) CreateEvent(
	token domain.Token,
	event domain.Event,
) (*domain.Token, error) {
	ctx := context.Background()

	oauth2Token := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    "Bearer",
	}

	client := c.oauth2Config.Client(ctx, oauth2Token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	cal, err := srv.Calendars.Get("primary").Do()
	if err != nil {
		return nil, err
	}

	fmt.Printf("User's default time zone: %s\n", cal.TimeZone)

	// Define event details.
	calEvent := &calendar.Event{
		Summary:     event.Title,
		Location:    event.Location,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.Start.DateTime,
			TimeZone: cal.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: event.End.DateTime,
			TimeZone: cal.TimeZone,
		},
	}

	// Insert the event into the user's primary calendar.
	createdEvent, err := srv.Events.Insert("primary", calEvent).Do()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Event created: %s\n", createdEvent.Id)

	// Return the token (if you need it for further requests).
	return &domain.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}, nil
}
