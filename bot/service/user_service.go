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
	"errors"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/storage"
)

func NewUserService(
	userRepository storage.UserRepository,
	tokenServices map[model.Provider]TokenService,
	timezoneService TimezoneService,
	linkedAccountsRepository storage.LinkedAccountRepository,
) *UserService {
	return &UserService{
		userRepository:           userRepository,
		tokenServices:            tokenServices,
		timezoneService:          timezoneService,
		linkedAccountsRepository: linkedAccountsRepository,
	}
}

type UserService struct {
	userRepository           storage.UserRepository
	tokenServices            map[model.Provider]TokenService
	timezoneService          TimezoneService
	linkedAccountsRepository storage.LinkedAccountRepository
}

func (s *UserService) GetUserByID(id int) (model.User, error) {
	return s.userRepository.GetByID(id)
}

func (s *UserService) GetUserByTelegramID(telegramID int64) (model.User, error) {
	return s.userRepository.GetByTelegramID(telegramID)
}

func (s *UserService) GetUserProfileByTelegramID(telegramID int64) (model.UserProfile, error) {
	user, err := s.GetUserByTelegramID(telegramID)
	if err != nil {
		return model.UserProfile{}, err
	}

	accounts, err := s.linkedAccountsRepository.GetByUserID(user.ID)
	if err != nil {
		return model.UserProfile{}, err
	}

	return model.UserProfile{
		User:           user,
		LinkedAccounts: accounts,
	}, nil
}

func (s *UserService) CreateUser(user *model.User) error {
	return s.userRepository.Save(user)
}

func (s *UserService) GetOAuth2Url(telegramID int64, callback func(error), provider model.Provider) (string, error) {
	user, err := s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return "", err
	} else if user.ID == 0 {
		return "", errors.New("user not found")
	}

	return s.tokenServices[provider].GetOAuth2URL(user.ID, callback)
}

func (s *UserService) LinkAccount(state string, provider model.Provider, code string) error {
	return s.tokenServices[provider].ExchangeCodeForToken(state, code)
}

func (s *UserService) UnlinkAccount(telegramID int64, provider model.Provider) (bool, error) {
	user, err := s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return false, err
	} else if user.ID == 0 {
		return false, errors.New("user not found")
	}

	return s.linkedAccountsRepository.DeleteByUserIDAndProvider(user.ID, provider)
}

func (s *UserService) UpdateUserTimeZone(telegramID int64, latitude float64, longitude float64) (string, error) {
	user, err := s.GetUserByTelegramID(telegramID)
	if err != nil {
		return "", err
	}

	timezone, err := s.timezoneService.GetTimezone(latitude, longitude)
	if err != nil {
		return "", err
	}

	user.Timezone = timezone.TimezoneId
	err = s.userRepository.Save(&user)
	if err != nil {
		return "", err
	}

	return timezone.TimezoneId, nil
}
