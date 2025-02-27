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
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Start       DateTime `json:"start"`
	End         DateTime `json:"end"`
	Location    string   `json:"location"`
	EventType   string   `json:"eventType"`
	DeepLink    string   `json:"deepLink"`
}

type DateTime struct {
	Timestamp time.Time `json:"timestamp"`
	TimeZone  string    `json:"timeZone"`
}

func (dt *DateTime) MarshalJSON() ([]byte, error) {
	type Alias DateTime
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(dt),
		Timestamp: dt.Timestamp.Format(time.DateTime),
	})
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	type Alias DateTime
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias: (*Alias)(dt),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	dt.Timestamp, _ = time.Parse(time.DateTime, aux.Timestamp)
	return nil
}
