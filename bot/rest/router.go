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

package rest

import (
	"net/http"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/service"
	"github.com/ivgag/schedulr/tgbot"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	tgBotConfig *tgbot.TelegramBotConfig,
	userService *service.UserService,
) *gin.Engine {
	router := gin.Default()

	router.GET("/oauth2callback/google", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		err := userService.LinkAccount(state, model.ProviderGoogle, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, tgBotConfig.URL)
	})

	return router
}
