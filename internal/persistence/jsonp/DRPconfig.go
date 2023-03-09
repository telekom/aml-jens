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

// Create a New DrPlayDataRateConfig
func NewDrPlayDataRateConfig(WarmupBeforeDrpMs float64, Frequency int, Scale float64, MinRateKbits float64) DrPlayDataRateConfig {
	return DrPlayDataRateConfig{
		WarmupBeforeDrpMs: &WarmupBeforeDrpMs,
		Frequency:         &Frequency,
		Scale:             &Scale,
		MinRateKbits:      &MinRateKbits,
	}
}

// Validate membervariables
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
