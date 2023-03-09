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
	data "github.com/telekom/aml-jens/cmd/drshow/internal/data/flowdata"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/shared"
)

type FlowChannels struct {
	UpdateText         chan shared.StrWithTextOpts
	UpdateLoad         chan data.DisplayDataDualT
	UpdateEcn          chan data.DisplayDataT
	UpdateDropp        chan data.DisplayDataT
	UpdateSojourn      chan data.DisplayDataT
	UpdateDelay        chan data.DisplayDataT
	RedrawUIWithLayout chan int32
	UpdateFlowDetails  chan *data.FlowT
	UpdateDataGraphs   chan *data.FlowT
	UpdateFlowList     chan *data.FlowManager
}

func (c *FlowChannels) Close() {
	close(c.UpdateText)
	close(c.UpdateLoad)
	close(c.UpdateEcn)
	close(c.UpdateDropp)
	close(c.UpdateSojourn)
	close(c.UpdateDelay)
	close(c.RedrawUIWithLayout)
	close(c.UpdateFlowDetails)
	close(c.UpdateDataGraphs)
	close(c.UpdateFlowList)

}

func NewFlowChannels() *FlowChannels {
	return &FlowChannels{
		UpdateText:         make(chan shared.StrWithTextOpts),
		UpdateLoad:         make(chan data.DisplayDataDualT),
		UpdateEcn:          make(chan data.DisplayDataT),
		UpdateDropp:        make(chan data.DisplayDataT),
		UpdateSojourn:      make(chan data.DisplayDataT),
		UpdateDelay:        make(chan data.DisplayDataT),
		RedrawUIWithLayout: make(chan int32),
		UpdateFlowDetails:  make(chan *data.FlowT),
		UpdateDataGraphs:   make(chan *data.FlowT),
		UpdateFlowList:     make(chan *data.FlowManager),
	}
}
