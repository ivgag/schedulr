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

package storage

import (
	"time"

	"github.com/ivgag/schedulr/model"
)

type LinkedAccount struct {
	ID           int
	UserID       int
	Provider     model.Provider
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

type LinkedAccountRepository interface {
	Save(account LinkedAccount) error
	GetByUserIDAndProvider(userID int, provider model.Provider) (LinkedAccount, error)
}
