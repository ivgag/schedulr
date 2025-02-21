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
	ExtractCalendarEvents(messages []model.TextMessage) (AiResponse[[]model.Event], model.Error)
	Provider() AIProvider
}

func extractCalendarEventsPrompt() string {
	return fmt.Sprintf(`
		You are an AI assistant that extracts structured event details from user input 
		(such as announcements, tickets, advertisements, or related content) and converts 
		them into a JSON array for creating calendar events (e.g., Google, Microsoft, Yandex).

		Tasks

		1. Extract Key Event Details
		• Title (required)
		• Description – Preserve all critical information (e.g., price, links, host’s name, 
			participants, rules, format). Keep these details as close to the original text as possible.
		• Start Date/Time (required)
		• End Date/Time (required if available; otherwise set a default).
		• Location (if provided).
		• Event Type – Must be one of: "event", "reminder", "meeting", "birthday", "holiday", "other".

		2. Resolve Relative Dates
		Use the provided reference date (e.g., "Today is %s") to convert relative expressions 
		like “tomorrow” or “next Friday” into absolute dates.

		3. Handle Incomplete Data
		• At a minimum, extract the title, start time, and end time.
		• If the end time is missing:
			– Use a known default duration for the event type,
			– Otherwise, assume a one-hour duration.

		4. Fallback Handling
		• If no event details are found, return an empty JSON array.
		• Provide a brief explanation of the result.

		Input Format
		The input may be single or multiple Telegram messages, 
		possibly including forwarded messages or bot commands. 
		Parse all text to identify event-related information.

		Output Format
		• Dates/times must follow the format: "YYYY-MM-DD HH:MM:SS".
		• Prices must be numeric or "free".
		• Links must be valid URLs.
		• The output must include a brief explanation of the result.
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
			sb.WriteString(fmt.Sprintf("The user's message: %s\n", msg.Text))
		case model.ForwardedMessage:
			sb.WriteString(fmt.Sprintf("Forwarded from %s: %s\n", msg.From, msg.Text))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

type AiResponse[T any] struct {
	Result      T      `json:"result"`
	Explanation string `json:"explanation"`
}

type EventSchema struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Location    string `json:"location"`
	EventType   string `json:"eventType"`
}
