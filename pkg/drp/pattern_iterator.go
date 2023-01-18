/*
 * aml-jens
 *
 * (C) 2023 Deutsche Telekom AG
 *
 * Deutsche Telekom AG and all other contributors /
 * copyright owners license this file to you under the Apache
 * License, Version 2.0 (the "License"); you may not use this
 * file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package drp

import (
	"github.com/telekom/aml-jens/internal/errortypes"
)

type DataRatePatternIterator struct {
	position int
	looping  bool
	operator int
	data     *[]float64
}

func NewDataRatePatternIterator(drp *DataRatePattern) *DataRatePatternIterator {
	return &DataRatePatternIterator{
		position: -1,
		looping:  false,
		operator: +1,
		data:     drp.data,
	}
}
func (s *DataRatePatternIterator) Value() float64 {
	if s.position < 0 {
		//TODO: Evaluate occurings
		INFO.Println("Tried to Value() @ -1")
		return (*s.data)[0]
	}
	return (*s.data)[s.position-s.operator]
}
func (s *DataRatePatternIterator) Next() (float64, error) {
	last_pos := s.position
	if s.position >= len(*s.data) {
		if s.looping {
			s.operator = -1
			last_pos -= 1
			s.position -= 1
		} else {
			return -1, &errortypes.IterableStopError{Msg: "end of DRP"}
		}
	} else {
		s.operator = +1
		last_pos += 1
		s.position += 1
	}
	s.position += int(s.operator)
	return (*s.data)[last_pos], nil
}
func (s *DataRatePatternIterator) SetLooping(endless bool) {
	s.looping = endless
}
