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

package channel

import (
	data "github.com/telekom/aml-jens/pkg/drp"
)

type DrpChannels struct {
	UpdateDrpPattern chan data.DataRatePattern
	UpdateDrpDetails chan data.DataRatePattern
	UpdateDrpList    chan data.DrpListT
}

func NewStaticDrpChannels() *DrpChannels {
	return &DrpChannels{
		UpdateDrpDetails: make(chan data.DataRatePattern),
		UpdateDrpList:    make(chan data.DrpListT),
	}
}
func NewDrpChannels() *DrpChannels {
	return &DrpChannels{
		UpdateDrpPattern: make(chan data.DataRatePattern),
		UpdateDrpDetails: make(chan data.DataRatePattern),
		UpdateDrpList:    make(chan data.DrpListT),
	}
}
func (s *DrpChannels) UpdateDrpGlobally(drp data.DataRatePattern) {
	s.UpdateDrpDetails <- drp
	if s.UpdateDrpPattern != nil {
		s.UpdateDrpPattern <- drp
	}
}
