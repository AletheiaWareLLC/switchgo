/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package switchgo

import (
	"log"
	"time"
)

type Switch struct {
	Name      string
	Timestamp uint64
	State     string
	Next      string
}

func (s *Switch) Switch(state string) {
	s.Timestamp = uint64(time.Now().UnixNano())
	s.State = state
	switch state {
	case "on":
		s.Next = "off"
	case "off":
		s.Next = "on"
	}
	log.Println("Switch", s)
	// TODO
}
