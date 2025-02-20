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

package model

import (
	"encoding/json"
	"time"
)

type ScheduledEvent struct {
	Event Event  `json:"event"`
	Link  string `json:"link"`
}

type Event struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Location    string    `json:"location"`
	EventType   string    `json:"eventType"`
}

func (e *Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		*Alias
		Start string `json:"stamp"`
		End   string `json:"end"`
	}{
		Alias: (*Alias)(e),
		Start: e.Start.Format(time.DateTime),
		End:   e.End.Format(time.DateTime),
	})
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		*Alias
		Start string `json:"stamp"`
		End   string `json:"end"`
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.Start, _ = time.Parse(time.DateTime, aux.Start)
	e.End, _ = time.Parse(time.DateTime, aux.End)
	return nil
}
