/*
 * Created on Mon Feb 17 2025
 *
 *  Copyright (c) 2025 Ivan Gagarkin
 * SPDX-License-Identifier: EPL-2.0
 *
 * Licensed under the Eclipse Public License - v 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.eclipse.org/legal/epl-2.0/
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/storage"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

var stateTokens = make(map[string]int)
var callbacks = make(map[string]func(error))

// NewGoogleTokenService creates a new TokenService.
func NewGoogleTokenService(
	config *GoogleConfig,
	linkedAccountsRepository storage.LinkedAccountRepository,
) *GoogleTokenService {
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{calendar.CalendarScope},
	}

	return &GoogleTokenService{
		oauth2Config:             oauth2Config,
		linkedAccountsRepository: linkedAccountsRepository,
	}
}

// GoogleTokenService encapsulates OAuth2 token logic.
type GoogleTokenService struct {
	oauth2Config             *oauth2.Config
	linkedAccountsRepository storage.LinkedAccountRepository
}

// GetOAuth2URL returns the URL to redirect users for Google OAuth2 consent.
func (s *GoogleTokenService) GetOAuth2URL(userID int, callback func(error)) (string, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	state := uuid.String()
	stateTokens[state] = userID
	callbacks[state] = callback

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
	delete(stateTokens, state)

	gToken, err := s.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	err = s.linkedAccountsRepository.Save(&model.LinkedAccount{
		UserID:       usedID,
		Provider:     model.ProviderGoogle,
		AccessToken:  gToken.AccessToken,
		RefreshToken: gToken.RefreshToken,
		Expiry:       gToken.Expiry.UTC(),
	})

	callback := callbacks[state]
	callback(err)
	delete(callbacks, state)

	return err
}

// ClientFromToken creates an HTTP client authenticated with the given token.
func (s *GoogleTokenService) ClientForUser(userID int) (*http.Client, error) {
	account, err := s.linkedAccountsRepository.GetByUserIDAndProvider(userID, model.ProviderGoogle)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(account.Expiry.UTC()) {
		tokenSource := s.oauth2Config.TokenSource(context.Background(), &oauth2.Token{
			RefreshToken: account.RefreshToken,
		})
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, err
		}

		account.AccessToken = newToken.AccessToken
		account.RefreshToken = newToken.RefreshToken
		account.Expiry = newToken.Expiry.UTC()

		err = s.linkedAccountsRepository.Save(&account)
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

type GoogleConfig struct {
	ApiKey       string `mapstructure:"api_key"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}
