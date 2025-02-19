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
