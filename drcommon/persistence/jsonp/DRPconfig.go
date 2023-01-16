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

import (
	"fmt"
)

type DrPlayDataRateConfig struct {
	WarmupBeforeDrpMs *float64 `json:",omitempty"`
	Frequency         *int     `json:",omitempty"`
	Scale             *float64 `json:",omitempty"`
	MinRateKbits      *float64 `json:",omitempty"`
}

func (s *DrPlayDataRateConfig) Equals(other DrPlayDataRateConfig) bool {
	return *s.WarmupBeforeDrpMs == *other.WarmupBeforeDrpMs &&
		*s.Frequency == *other.Frequency &&
		*s.Scale == *other.Scale &&
		*s.MinRateKbits == *other.MinRateKbits
}

func (s *DrPlayDataRateConfig) String() string {
	return fmt.Sprintf(`DrPlayDRPconfig{
WarmupBeforeDrpMs: %f,
Frequency: %d,
Scale: %f,
MinRateKbits: %f
}`, *s.WarmupBeforeDrpMs, *s.Frequency, *s.Scale, *s.MinRateKbits)
}
func NewDrPlayDataRateConfig(WarmupBeforeDrpMs float64, Frequency *int, Scale *float64, MinRateKbits float64) DrPlayDataRateConfig {
	return DrPlayDataRateConfig{
		WarmupBeforeDrpMs: &WarmupBeforeDrpMs,
		Frequency:         Frequency,
		Scale:             Scale,
		MinRateKbits:      &MinRateKbits,
	}
}
func NewDrPlayDataRateConfigD(WarmupBeforeDrpMs float64, Frequency int, Scale float64, MinRateKbits float64) DrPlayDataRateConfig {
	return DrPlayDataRateConfig{
		WarmupBeforeDrpMs: &WarmupBeforeDrpMs,
		Frequency:         &Frequency,
		Scale:             &Scale,
		MinRateKbits:      &MinRateKbits,
	}
}

//Update updated the config with the values supplied in other
func (self *DrPlayDataRateConfig) Update(other *DrPlayDataRateConfig) {
	if other == nil {
		return
	}
	if self == nil {
		self = &DrPlayDataRateConfig{}
	}
	if other.WarmupBeforeDrpMs != nil {
		if self.WarmupBeforeDrpMs == nil {
			self.WarmupBeforeDrpMs = new(float64)
		}
		*self.WarmupBeforeDrpMs = *other.WarmupBeforeDrpMs
	}
	if other.Frequency != nil {
		if self.Frequency == nil {
			self.Frequency = new(int)
		}
		*self.Frequency = *other.Frequency
	}
	if other.Scale != nil {
		if self.Scale == nil {
			self.Scale = new(float64)
		}
		*self.Scale = *other.Scale

	}
	if other.MinRateKbits != nil {
		if self.MinRateKbits == nil {
			self.MinRateKbits = new(float64)
		}
		*self.MinRateKbits = *other.MinRateKbits

	}
}

//Update updated the config with the values supplied in other
//Only if self has no value set
func (self *DrPlayDataRateConfig) UpdateWhereNil(other DrPlayDataRateConfig) {

	if self == nil {
		self = &DrPlayDataRateConfig{}
	}
	if other.WarmupBeforeDrpMs != nil && self.WarmupBeforeDrpMs == nil {
		self.WarmupBeforeDrpMs = new(float64)
		*self.WarmupBeforeDrpMs = *other.WarmupBeforeDrpMs
	}
	if other.Frequency != nil && self.Frequency == nil {
		self.Frequency = new(int)
		*self.Frequency = *other.Frequency
	}
	if other.Scale != nil && self.Scale == nil {
		self.Scale = new(float64)
		*self.Scale = *other.Scale
	}
	if other.MinRateKbits != nil && self.MinRateKbits == nil {
		self.MinRateKbits = new(float64)
		*self.MinRateKbits = *other.MinRateKbits
	}
}
func (bdrp *DrPlayDataRateConfig) Validate() error {
	E := func(s string) error { return fmt.Errorf("BenchmarkDrplaySetting: %s", s) }
	if bdrp.Frequency != nil && (*bdrp.Frequency > 100 || *bdrp.Frequency < 0) {
		return E("Invalid Frequency range [0-100]")
	}
	if bdrp.Scale != nil && *bdrp.Scale < 0.1 {
		return E("Scale must be >= 0.1")
	}
	if bdrp.MinRateKbits != nil && *bdrp.MinRateKbits < 0 {
		return E("MinRateKbits can't be less than 0")
	}
	if bdrp.WarmupBeforeDrpMs != nil && *bdrp.WarmupBeforeDrpMs < 0 {
		return E("WarmupBeforeDrpMs can't be less than 0")
	}
	return nil
}
