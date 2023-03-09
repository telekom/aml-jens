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

type FlowManagerEventHanderlerT struct {
	NewFlow          func(man *FlowManager) error
	NewDataInFlow    func(flow *FlowT) error
	NewFlowSelected  func(man *FlowManager) error
	UpdateFlow       func(man *FlowManager, index int) error
	Exit             func(man *FlowManager)
	OnFirstFlowAdded func(man *FlowManager)
	OnFailureOrEnd   func(code uint8, text string)
	OnEndOfDrpplayer func()
}
