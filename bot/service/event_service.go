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
	"github.com/ivgag/schedulr/model"
)

func NewEventService(
	aiService AIService,
	userService UserService,
	clanedarServices []CalendarService,
) *EventService {
	calServices := make(map[model.Provider]CalendarService)
	for _, service := range clanedarServices {
		calServices[service.Provider()] = service
	}

	return &EventService{
		aiService:        aiService,
		userService:      userService,
		calendarServices: calServices,
	}
}

type EventService struct {
	aiService        AIService
	userService      UserService
	calendarServices map[model.Provider]CalendarService
}

func (s *EventService) CreateEventsFromUserMessage(
	telegramID int64,
	messages *[]model.TextMessage,
) (*[]model.ScheduledEvent, error) {

	profile, err := s.userService.GetUserProfileByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	events, err := s.aiService.ExtractCalendarEvents(profile.User.TimeZone, messages)
	if err != nil {
		return nil, err
	}

	var calSrv CalendarService
	for provider, service := range s.calendarServices {
		if profile.LinkedAccount(provider) != nil {
			calSrv = service
			break
		}
	}

	var scheduledEvents []model.ScheduledEvent = make([]model.ScheduledEvent, len(*events))
	for i, event := range *events {
		var scheduledEvent model.ScheduledEvent
		var err error

		if calSrv != nil {
			scheduledEvent, err = calSrv.CreateEvent(profile.User.ID, &event)
		} else {
			scheduledEvent = model.ScheduledEvent{
				Event: event,
				Link:  "",
			}
		}

		if err != nil {
			return nil, err
		}
		scheduledEvents[i] = scheduledEvent
	}

	return &scheduledEvents, nil
}
