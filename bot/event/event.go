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

package event

import (
	"github.com/ivgag/schedulr/model"
)

type EventType string

const (
	TokenRefreshed EventType = "TokenRefreshed"
)

type BaseEvent interface {
	Type() EventType
	Payload() any
}

// Event is a marker interface for all events.
type Event[P any] interface {
	BaseEvent
	TypedPayload() P
}

// EventHandler defines a handler that can process events.
type EventHandler[P any] interface {
	Handle(event BaseEvent)
	Type() EventType
}

// EventDispatcher defines a simple interface to register handlers and dispatch events.
type EventDispatcher interface {
	Register(handler EventHandler[any])
	Dispatch(event BaseEvent)
}

// SimpleEventDispatcher is a basic implementation of the EventDispatcher.
type SimpleEventDispatcher struct {
	handlers map[EventType][]EventHandler[any]
}

// NewSimpleEventDispatcher creates a new dispatcher.
func NewSimpleEventDispatcher() *SimpleEventDispatcher {
	return &SimpleEventDispatcher{
		handlers: make(map[EventType][]EventHandler[any]),
	}
}

// Register adds a new event handler.
func (d *SimpleEventDispatcher) Register(handler EventHandler[any]) {
	if _, ok := d.handlers[handler.Type()]; !ok {
		d.handlers[handler.Type()] = make([]EventHandler[any], 0)
	}

	d.handlers[handler.Type()] = append(d.handlers[handler.Type()], handler)
}

// Dispatch sends the event to all registered handlers.
func (d *SimpleEventDispatcher) Dispatch(event BaseEvent) {
	for _, handler := range d.handlers[event.Type()] {
		handler.Handle(event)
	}
}

type OAuth2TokenEvent struct {
	Token model.Token
}

func (e OAuth2TokenEvent) Type() EventType {
	return TokenRefreshed
}

// Satisfies BaseEvent.
func (e OAuth2TokenEvent) Payload() any {
	return e.Token
}

// Provides type safety.
func (e OAuth2TokenEvent) TypedPayload() model.Token {
	return e.Token
}
