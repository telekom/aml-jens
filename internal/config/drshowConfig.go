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

package config

type ConfigFlowPlotScrollMode uint8

const (
	Scrolling = iota
	Scaling
)

type showConfigFlowEligibilityEntryT struct {
	AdditionalCount   int // No filter until this+previous num of Flows is met
	TimePastMaxInSec  int
	MinimumFlowLength int
}
type showConfigFlowEligibilityT struct {
	Level0Until int // No filter until this num of Flows is met
	Level1      showConfigFlowEligibilityEntryT
	Level2      showConfigFlowEligibilityEntryT
	Level3      showConfigFlowEligibilityEntryT
}
type showConfig struct {
	PlotScrollMode          ConfigFlowPlotScrollMode
	WaitTimeForSamplesInSec int32
	ExportPathPrefix        string
	FlowEligibility         showConfigFlowEligibilityT
}
