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

package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivgag/schedulr/model"
)

type AIProvider string

const (
	ProviderOpenAI   AIProvider = "OpenAI"
	ProviderDeepSeek AIProvider = "DeepSeek"
)

type AI interface {
	ExtractCalendarEvents(message *model.TextMessage) ([]model.Event, model.Error)
	Provider() AIProvider
}

func extractCalendarEventsPrompt() string {
	return fmt.Sprintf(`
	You are a calendar assistant that extracts event details from user input—such as announcements, 
	tickets, ads, or related content—and converts them into a JSON array for creating calendar events 
	(e.g., in Google, Microsoft, Yandex).

	Tasks
		1.	Extract Key Event Details:
			•	Title
			•	Description
			•	Start date/time
			•	End date/time
			•	Location
			•	Event type
			•	Price
			•   Links
			• 	Host’s name
		2.	Resolve Relative Dates:
			Use the provided reference date. For example:
			"Today is %s"
		3.	Handling Incomplete Data:
			•	At a minimum, extract the title, start time, and end time.
			•	If there is no information about the end time, use a default duration:
			•	If you know the usual duration for the event type, use that; otherwise, assume a one-hour duration.
		4.	Fallback:
			•	If no event details are found in the input, return an empty JSON array.
			•   Write a brief explanation of the result.
		5.	Output Requirement:

	Input Format
		•	The input may include single or multiple Telegram messages.
		•	Messages can be:
		•	Forwarded from an events channel in the user’s city.
		•	A forwarded conversation between users.
		•	A combination of forwarded messages and commands to the bot.
		•	Parse all incoming text to identify any event-related information.

	Output Format
		•	Ensure that all extracted dates and times are formatted according to "YYYY-MM-DD HH:MM:SS".
		•	Ensure that the event type is one of the following: "event", "reminder", "meeting", "birthday", "holiday", "other".
		•	Ensure that the price is a number or "free".
		•	Ensure that the links are valid URLs.
		•	Ensure that the output includes an explanation of the result.
	`,
		time.Now().Format(time.DateTime),
	)
}

func removeJsonFormattingMarkers(text string) string {
	// Remove formatting markers (```json and trailing backticks)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")
	return text
}

// messagesToText converts an array of TextMessages into a single string.
// It uses formatMessageText to include entity markers correctly.
func messagesToText(messages []model.TextMessage) string {
	var sb strings.Builder

	for _, msg := range messages {
		switch msg.MessageType {
		case model.UserMessage:
			sb.WriteString(fmt.Sprintf("%s: %s\n", msg.From, msg.Text))
		case model.ForwardedMessage:
			sb.WriteString(fmt.Sprintf("Forwarded from %s: %s\n", msg.From, msg.Text))
		}
	}

	return sb.String()
}

type AiResponseDto[T any] struct {
	Result      T      `json:"result"`
	Explanation string `json:"explanation"`
}

type AiEventDto struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Location    string `json:"location"`
	EventType   string `json:"eventType"`
}

func (e *AiEventDto) ToModelEvent() (model.Event, error) {
	start, err := time.Parse(time.DateTime, e.Start)
	end, err := time.Parse(time.DateTime, e.End)

	return model.Event{
		Title:       e.Title,
		Description: e.Description,
		Start: model.TimeStamp{
			DateTime: start,
		},
		End: model.TimeStamp{
			DateTime: end,
		},
		Location:  e.Location,
		EventType: e.EventType,
	}, err
}
