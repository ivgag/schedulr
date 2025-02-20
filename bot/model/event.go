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
	"strings"
	"time"
)

type Event struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Start       TimeStamp `json:"start"`
	End         TimeStamp `json:"end"`
	Location    string    `json:"location"`
	EventType   string    `json:"eventType"`
	Link        string    `json:"link"`
}

type TimeStamp struct {
	DateTime time.Time `json:"dateTime"`
	TimeZone string    `json:"timeZone"`
}

type DateTime time.Time

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(dt).Format(time.DateTime) + `"`), nil
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "" || s == "null" {
		*dt = DateTime(time.Time{})
		return nil
	}
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return err
	}
	*dt = DateTime(t)
	return nil
}

func (dt DateTime) String() string {
	return time.Time(dt).Format(time.DateTime)
}
