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

package pipe

import (
	"context"

	coms "github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	flowWidgets "github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/flow"

	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

type uiSettings struct {
	cb_flow_listInc button.CallbackFn
	cb_flow_listDec button.CallbackFn
}

type comWids struct {
	FlowDetails *text.Text
	FlowList    *text.Text
	BtnNext     *button.Button
	BtnPrev     *button.Button
	InfoBox     *text.Text
}
type detWids struct {
	LoadChart    *linechart.LineChart
	EcnChart     *linechart.LineChart
	DroppChart   *linechart.LineChart
	SojournChart *linechart.LineChart
	DelayChart   *linechart.LineChart
}

func NewDetWids(ctx context.Context, c *coms.FlowChannels, flowId int32) detWids {
	Load, err := flowWidgets.NewFlowDataLineChartCombinationLoadCapacity(ctx, c.UpdateLoad, flowId)
	if err != nil {
		panic(err)
	}

	Ecn, err := flowWidgets.NewFlowDataLineChartPerc(ctx, c.UpdateEcn, flowId)
	if err != nil {
		panic(err)
	}
	Dropp, err := flowWidgets.NewFlowDataLineChartAbs(ctx, c.UpdateDropp, flowId)
	if err != nil {
		panic(err)
	}
	Sojourn, err := flowWidgets.NewFlowDataLineChartMs(ctx, c.UpdateSojourn, flowId)
	if err != nil {
		panic(err)
	}
	Delay, err := flowWidgets.NewFlowDataLineChart(ctx, c.UpdateDelay, flowId)
	if err != nil {
		panic(err)
	}
	return detWids{
		LoadChart:    Load,
		EcnChart:     Ecn,
		DroppChart:   Dropp,
		SojournChart: Sojourn,
		DelayChart:   Delay,
	}
}

type wids struct {
	//G detWids
	I comWids
	M map[int32]detWids
}
