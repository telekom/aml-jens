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
	"time"

	"github.com/telekom/aml-jens/internal/config"
)

func (self *FlowManager) isFlowEligible(flow *FlowT) bool {
	if !flow.hasCombinedLoadGreater1() {
		return false
	}
	cfg := config.ShowCfg().FlowEligibility
	isFromLastNSecsAndHasMoreThen3 := func(flow *FlowT, n int) bool {
		if flow.D.Length() == 0 { // How is this possible ? Maybe some Mutex is not set correctly
			return true // TODO: Fix this weird behavior! ^
		}
		return (*flow.D.TimeStamp)[flow.D.Length()-1] > float64(time.Now().UnixMilli()-int64(n*1000)) && flow.D.Length() > 3
	}
	flowCountThreshold := cfg.Level0Until
	if len(self.Flows) < flowCountThreshold { // Just 1 Row
		return true
	}
	selected, err := self.GetSelectedFlow()
	if err != nil && selected.Equals(flow) {
		return true
	}
	flowLength := flow.D.Length()
	flowCountThreshold += cfg.Level1.AdditionalCount
	if len(self.Flows) < flowCountThreshold { // 2 Rows
		return flowLength > cfg.Level1.MinimumFlowLength || isFromLastNSecsAndHasMoreThen3(flow, cfg.Level1.TimePastMaxInSec)
	}
	flowCountThreshold += cfg.Level2.AdditionalCount
	if len(self.Flows) < flowCountThreshold { // 3 Rows
		return flowLength > cfg.Level2.MinimumFlowLength || isFromLastNSecsAndHasMoreThen3(flow, cfg.Level2.TimePastMaxInSec)
	}

	return flowLength > cfg.Level3.MinimumFlowLength || (isFromLastNSecsAndHasMoreThen3(flow, cfg.Level3.TimePastMaxInSec))
}

type relativeIndexMatchType uint8

const (
	exactMatch relativeIndexMatchType = iota
	previousMatch
	nextMatch
)

type relativeIndexOfSelectedType struct {
	Index     int
	MatchType relativeIndexMatchType
}

func (man *FlowManager) EligibleFLows() (relativeIndexOfSelectedT relativeIndexOfSelectedType, absoluteIndexes []int, eligibleFlows []*FlowT) {
	amt := len(man.Flows)
	relativeIndexOfSelectedT = relativeIndexOfSelectedType{
		Index:     -1,
		MatchType: exactMatch,
	}
	absoluteIndexes = make([]int, 0, amt)
	eligibleFlows = make([]*FlowT, 0, amt)
	for i := 0; i < len(man.Flows); i++ {
		if man.isFlowEligible(man.Flows[i]) {
			absoluteIndexes = append(absoluteIndexes, i)
			eligibleFlows = append(eligibleFlows, man.Flows[i])
			if i == man.selectedFlowIndex {
				relativeIndexOfSelectedT.Index = len(absoluteIndexes) - 1
			}
			if i > man.selectedFlowIndex && relativeIndexOfSelectedT.Index == -1 {
				relativeIndexOfSelectedT.Index = len(absoluteIndexes) - 1
				relativeIndexOfSelectedT.MatchType = nextMatch
			}
		}

	}
	//log.Printf("%v -- %v %d --> %d\n", absoluteIndexes, eligableFlows, relativeIndexOfSelected, man.selectedFlowIndex)
	return relativeIndexOfSelectedT, absoluteIndexes, eligibleFlows

}
