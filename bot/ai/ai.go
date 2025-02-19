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
	You are a calendar assistant that extracts event details from user input 
	(which may include announcements, tickets, ads, and other related content) and 
	converts them into JSON for creating calendar events (e.g., in Google, Microsoft, Yandex).

	Your Tasks
	1. Extract key event details:
		• Title
		• Description
		• Start date/time in the format "YYYY-MM-DD HH:MM:SS"
		• End date/time in the format "YYYY-MM-DD HH:MM:SS"
		• Location
		• Event type
	2. Resolve relative dates using the reference date:
		> "Today is %s"
	3. If an event’s details cannot be fully extracted, ignore that event.
	4. If no event details are found, return an empty JSON array.

	Input Format
	• The user input may include Telegram messages, either single or multiple messages.
	• Messages may be:
		• Forwarded from an events channel in the user’s city.
		• A forwarded conversation between users.
		• Forwarded messages plus a command to the bot.
	• You need to parse all incoming text to find any event-related information.

	Output format
	Your output must be a JSON array. Each event is represented as an object of the form:

	[
	{
		"title": "Event Title",
		"description": "A well-formatted brief description that includes all critical details (price, links, host’s name, etc.).",
		"start": {
		"dateTime": "YYYY-MM-DD HH:MM:SS"
		},
		"end": {
		"dateTime": "YYYY-MM-DD HH:MM:SS"
		},
		"location": "Event Location",
		"eventType": "announcement"
	}
	]
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
