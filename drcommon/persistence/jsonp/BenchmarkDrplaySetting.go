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

type DrplaySetting struct {
	TC  *DrPlayTrafficControlConfig `json:",omitempty"`
	DRP *DrPlayDataRateConfig       `json:",omitempty"`
}

func (bdrp *DrplaySetting) Validate() error {
	if bdrp.TC != nil {
		if err := bdrp.TC.Validate(); err != nil {
			return err
		}
	}
	if bdrp.DRP != nil {
		if err := bdrp.DRP.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func NewDrplaySetting(freq int, scale float64, minrate float64) DrplaySetting {
	return DrplaySetting{
		TC: nil,
		DRP: &DrPlayDataRateConfig{
			Frequency:    &freq,
			Scale:        &scale,
			MinRateKbits: &minrate,
		}}
}
