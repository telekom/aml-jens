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

package jsonp

import "fmt"

type DrPlayTrafficControlConfig struct {
	Markfree            *int32  `json:"Markfree,omitempty"`
	Markfull            *int32  `json:"Markfull,omitempty"`
	Extralatency        *int32  `json:"Extralatency,omitempty"`
	L4sEnablePreMarking *bool   `json:"L4sEnablePreMarking,omitempty"`
	SignalDrpStart      *bool   `json:"SignalDrpStart,omitempty"`
	Queuesizepackets    *uint64 `json:"Queuesizepackets,omitempty"`
}

func (s *DrPlayTrafficControlConfig) Equals(other DrPlayTrafficControlConfig) bool {
	return *s.Markfree == *other.Markfree &&
		*s.Markfull == *other.Markfull &&
		*s.Extralatency == *other.Extralatency &&
		*s.L4sEnablePreMarking == *other.L4sEnablePreMarking &&
		*s.SignalDrpStart == *other.SignalDrpStart
}
func (s *DrPlayTrafficControlConfig) String() string {
	Markfree := "nil"
	Markfull := "nil"
	Extralatency := "nil"
	L4sEnablePreMarking := "nil"
	SignalDrpStart := "nil"
	Queuesizepackets := "nil"

	if s.Markfree != nil {
		Markfree = fmt.Sprintf("%d", *s.Markfree)
	}
	if s.Markfull != nil {
		Markfull = fmt.Sprintf("%d", *s.Markfull)
	}
	if s.Extralatency != nil {
		Extralatency = fmt.Sprintf("%d", *s.Extralatency)
	}
	if s.L4sEnablePreMarking != nil {
		L4sEnablePreMarking = fmt.Sprintf("%t", *s.L4sEnablePreMarking)
	}
	if s.SignalDrpStart != nil {
		SignalDrpStart = fmt.Sprintf("%t", *s.SignalDrpStart)
	}
	if s.Queuesizepackets != nil {
		Queuesizepackets = fmt.Sprintf("%d", *s.Queuesizepackets)
	}

	return fmt.Sprintf(`DrPlayTCconfig{
Mfree: %s,
Mfull: %s,
Extralatency: %s,
L4sPreM: %s,
SignalStart: %s,
Queuesizepackets: %s
}`, Markfree, Markfull, Extralatency, L4sEnablePreMarking, SignalDrpStart, Queuesizepackets)
}

func NewDrPlayTrafficControlConfig(markfree int32, markfull int32, extralatency int32, enablepremarking bool, signaldrpstart bool) DrPlayTrafficControlConfig {
	return DrPlayTrafficControlConfig{
		Markfree:            &markfree,
		Markfull:            &markfull,
		Extralatency:        &extralatency,
		L4sEnablePreMarking: &enablepremarking,
		SignalDrpStart:      &signaldrpstart,
	}
}

// Update updated the config with the values supplied in other
// Only if self has no value set
func (self *DrPlayTrafficControlConfig) UpdateWhereNil(other DrPlayTrafficControlConfig) {
	if self == nil {
		self = &DrPlayTrafficControlConfig{}
	}
	if other.Markfree != nil && self.Markfree == nil {
		self.Markfree = new(int32)
		*self.Markfree = *other.Markfree
	}
	if other.Markfull != nil && self.Markfull == nil {
		self.Markfull = new(int32)
		*self.Markfull = *other.Markfull
	}
	if other.Extralatency != nil && self.Extralatency == nil {
		self.Extralatency = new(int32)
		*self.Extralatency = *other.Extralatency
	}
	if other.L4sEnablePreMarking != nil && self.L4sEnablePreMarking == nil {
		self.L4sEnablePreMarking = new(bool)
		*self.L4sEnablePreMarking = *other.L4sEnablePreMarking
	}
	if other.SignalDrpStart != nil && self.SignalDrpStart == nil {
		self.SignalDrpStart = new(bool)
		*self.SignalDrpStart = *other.SignalDrpStart
	}
}

func (self *DrPlayTrafficControlConfig) Update(other *DrPlayTrafficControlConfig) {
	if other == nil {
		return
	}
	if self == nil {
		self = &DrPlayTrafficControlConfig{}
	}
	if other.Markfree != nil {
		self.Markfree = other.Markfree
	}
	if other.Markfull != nil {
		self.Markfull = other.Markfull
	}
	if other.Extralatency != nil {
		self.Extralatency = other.Extralatency
	}
	if other.L4sEnablePreMarking != nil {
		self.L4sEnablePreMarking = other.L4sEnablePreMarking
	}
	if other.SignalDrpStart != nil {
		self.SignalDrpStart = other.SignalDrpStart
	}
}

func (tcSet *DrPlayTrafficControlConfig) Validate() error {
	E := func(s string) error { return (fmt.Errorf("tcDrPlaySetting: %s", s)) }
	if (tcSet.Markfree != nil && tcSet.Markfull != nil) && *tcSet.Markfree > *tcSet.Markfull {
		return E("Markfree can not be greater or equal to Markfull")
	}

	return nil
}
