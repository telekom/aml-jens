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
	"github.com/telekom/aml-jens/internal/util"
)

type DataRatePatternIterator struct {
	looping  bool
	operator int
	data     *[]float64
	position int
	value    float64
}

func NewDataRatePatternIterator() *DataRatePatternIterator {
	return &DataRatePatternIterator{
		looping:  false,
		operator: +1,
		position: -1,
	}
}

// updates internal data pointer as well as position and value
func (s *DataRatePatternIterator) updateAndReset(drp *[]float64) {
	s.data = drp
	s.Reset()
}

// Resets positon and value to start
func (s *DataRatePatternIterator) Reset() {
	s.position = -1
	s.value = (*s.data)[0]
}

// Retunrs last value
func (s *DataRatePatternIterator) Value() float64 {
	return s.value
}

// Get next Value
func (s *DataRatePatternIterator) Next() (float64, error) {
	switch max_i := len(*s.data); {
	case s.position == -1:
		s.position += s.operator * 2
	case s.position >= max_i || s.position <= 0:
		at_max := s.position >= max_i
		//We are at the end of a cycle
		if !s.looping {
			return 0, &errortypes.IterableStopError{}
		}
		//reverse driection
		s.operator *= -1
		if at_max {
			//double last value and continue at last -1
			s.position += s.operator * 2
		} else /* at_min*/ {
			//start back up at 0- doubling it
			s.position = -1
			s.value = (*s.data)[0]
		}

	default:
		s.position += s.operator
		index := util.MinInt(max_i-1, util.AbsInt(s.position-s.operator))
		s.value = (*s.data)[index]
	}
	return s.value, nil
}

// Turns on / off looping-mode
func (s *DataRatePatternIterator) SetLooping(endless bool) {
	s.looping = endless
}

// returns if looping is on
func (s *DataRatePatternIterator) IsLooping() bool {
	return s.looping
}

// Sets the Iterator to a Doen state.
// Subsequent calls to Next will raise an IterableStopError
func (s *DataRatePatternIterator) SetDone() {
	s.SetLooping(false)
	s.operator = +1
	s.position = len(*s.data)
}
