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

package flowdata

import (
	"sync"
	"time"
)

type FlowDatapointsT struct {
	TimeStamp       *[]float64
	Sojourn         *[]float64
	Load            *[]float64
	Ecn             *[]float64
	Dropp           *[]float64
	Delay           *[]float64
	Capacity        *[]float64
	mutex           sync.Mutex
	length          int
	lengthStateID   uint64
	stateID         uint64
	hasLoadGreater0 bool
	HasLoadGreater1 bool
}

func NewFlowDataPoints() *FlowDatapointsT {
	TimeStamp := make([]float64, 0, 8192)
	Sojourn := make([]float64, 0, 8192)
	Load := make([]float64, 0, 8192)
	Ecn := make([]float64, 0, 8192)
	Dropp := make([]float64, 0, 8192)
	Delay := make([]float64, 0, 8192)
	Capacity := make([]float64, 0, 8192)
	return &FlowDatapointsT{
		TimeStamp:       &TimeStamp,
		Sojourn:         &Sojourn,
		Load:            &Load,
		Ecn:             &Ecn,
		Dropp:           &Dropp,
		Delay:           &Delay,
		Capacity:        &Capacity,
		stateID:         0,
		hasLoadGreater0: false,
		HasLoadGreater1: false,
	}
}

func (s *FlowDatapointsT) Append(ts float64, sj float64, ld float64, ec float64, dr float64, cap float64) {
	s.mutex.Lock()
	*s.TimeStamp = append(*s.TimeStamp, ts)
	*s.Sojourn = append(*s.Sojourn, sj)
	*s.Load = append(*s.Load, ld)
	*s.Ecn = append(*s.Ecn, ec)
	*s.Dropp = append(*s.Dropp, dr)
	*s.Delay = append(*s.Delay, float64(time.Now().UnixMilli()-int64(ts)))
	*s.Capacity = append(*s.Capacity, cap)
	s.stateID++
	if !s.HasLoadGreater1 && ld > 0 {
		if !s.hasLoadGreater0 {
			s.hasLoadGreater0 = true
		} else {
			s.HasLoadGreater1 = true
		}
	}

	s.mutex.Unlock()
}
func (s *FlowDatapointsT) Length() int {
	if s.lengthStateID == s.stateID {
		return s.length
	}
	s.mutex.Lock()
	s.length = s.calcLength()
	s.mutex.Unlock()
	return s.length
}
func (s *FlowDatapointsT) calcLength() int {
	length := len(*s.TimeStamp)
	if !(length == len(*s.Sojourn) &&
		len(*s.Load) == len(*s.Ecn) &&
		len(*s.Ecn) == len(*s.Sojourn) &&
		len(*s.Sojourn) == len(*s.Dropp) &&
		len(*s.Dropp) == len(*s.Delay)) {
		FATAL.Exitln("FlowDatapoints are not in SYNC!")
	}
	return length
}
