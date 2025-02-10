package google

import (
	"context"

	"github.com/ivgag/schedulr/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
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

func (c *GoogleClient) GetCalendarConnectionUrl(oauthStateString string) string {
	return c.oauth2Config.AuthCodeURL(
		oauthStateString,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"), // Always ask for refresh token.
	)
}

func (c *GoogleClient) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return c.oauth2Config.Exchange(context.Background(), code)
}
