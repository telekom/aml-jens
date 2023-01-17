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
	"errors"
	"fmt"
	"time"

	coms "github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/flowdata"
	flowWidgets "github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/flow"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/shared"
	"github.com/telekom/aml-jens/internal/util"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func NewUI(ctx context.Context, t terminalapi.Terminal, c *container.Container, set *uiSettings, C *coms.FlowChannels, man *flowdata.FlowManager) (*wids, error) {
	txt, err := shared.NewTextWithOptsBoxAppendTopLast2(ctx, t, C.UpdateText)
	if err != nil {
		return nil, err
	}

	flowDetails, err := flowWidgets.NewFlowDetailsTextBox(ctx, t, C)
	if err != nil {
		return nil, err
	}

	flowList, err := flowWidgets.NewFlowListTextBox(ctx, t, C.UpdateFlowList)
	if err != nil {
		return nil, err
	}
	flowBtnInc, err := flowWidgets.NewFlowListButton(ctx, t, flowWidgets.FlowLBtnUP(), func() error {

		err := set.cb_flow_listInc()
		//C.RedrawUIWithLayout <- man.Flows[man.SelectedFlowIndex].FlowId
		return err
	})
	if err != nil {
		return nil, err
	}
	flowBtnDec, err := flowWidgets.NewFlowListButton(ctx, t, flowWidgets.FlowLBtnDown(), func() error {
		err := set.cb_flow_listDec()
		//C.RedrawUIWithLayout <- man.Flows[man.SelectedFlowIndex].FlowId
		return err
	})
	if err != nil {
		return nil, err
	}

	s := &wids{
		I: comWids{
			FlowDetails: flowDetails,
			FlowList:    flowList,
			BtnNext:     flowBtnInc,
			BtnPrev:     flowBtnDec,
			InfoBox:     txt,
		},
		M: map[int32]detWids{},
	}
	C.UpdateText <- shared.NewStrTextWriteOps("To quit press <ESC> || <C-C> || <q>\n" +
		"Maximize window to view all graphs (Load, Sojourntime, ECN%, Drops)\n" +
		"Flows: Navigate the list with <ArrowUp> and <ArrowDown>\n" +
		"       The selected flow is highlighted and its details are shown.\n" +
		"       By pressing e the currently selected flow is exported.\n" +
		"Plots: Use mouse clicks and selection to zoom in. Mousewheel can also be used to zoom.")
	lastUpdate := int64(0)
	go func() {
		for {
			select {
			case flow := <-C.UpdateDataGraphs:
				man.Mutex.Lock()
				if s.M[flow.FlowId] == (detWids{}) {
					s.M[flow.FlowId] = NewDetWids(ctx, C, flow.FlowId)
				}
				C.UpdateLoad <- flowdata.DisplayDataDualT{Name: fmt.Sprint(flow.FlowId), FlowId: flow.FlowId, Data: flow.D.Load, Color: flow.Color(), DataExtra: flow.D.Capacity}
				C.UpdateDelay <- flowdata.DisplayDataT{Name: fmt.Sprint(flow.FlowId), FlowId: flow.FlowId, Data: flow.D.Delay, Color: flow.Color()}
				C.UpdateDropp <- flowdata.DisplayDataT{Name: fmt.Sprint(flow.FlowId), FlowId: flow.FlowId, Data: flow.D.Dropp, Color: flow.Color()}
				C.UpdateEcn <- flowdata.DisplayDataT{Name: fmt.Sprint(flow.FlowId), FlowId: flow.FlowId, Data: flow.D.Ecn, Color: flow.Color()}
				C.UpdateSojourn <- flowdata.DisplayDataT{Name: fmt.Sprint(flow.FlowId), FlowId: flow.FlowId, Data: flow.D.Sojourn, Color: flow.Color()}
				man.Mutex.Unlock()
				now := time.Now().UnixMilli()
				if now-lastUpdate >= 100 {
					lastUpdate = now
					go func() {
						selectedFlow, err := man.GetSelectedFlow()
						if err != nil {
							return
						}
						C.UpdateFlowList <- man
						if flow.Equals(selectedFlow) {
							C.UpdateFlowDetails <- selectedFlow
						}

					}()
				}

				//				C.UpdateText <- fmt.Sprintf("Updating Flow[%d] (%d)", flow.FlowId, len(*flow.D.Load))

			case <-ctx.Done():
				return
			}
		}
	}()

	return s, nil
}

const layoutCombined int32 = -1

type _heightConfig struct {
	topPart        int
	bottomPart     int
	bottomSubParts [4]int
}

func (h *_heightConfig) str() string {
	return fmt.Sprintf("\ntopPart:%d\nbotPart:%d\n\t%v", h.topPart, h.bottomPart, h.bottomSubParts)
}
func (h *_heightConfig) equals(other _heightConfig) bool {
	return h.topPart == other.topPart && h.bottomPart == other.bottomPart
}

func plotHeights(height int) _heightConfig {
	//log.Printf("Redrawing in terminal of size %dx%d", t.Size().X, t.Size().Y)
	const TOP_HEIGHT int = 9
	const MIN_AVG_BOT_HEIGHT int = 7
	const COUNT_PLTS int = 4
	const PRIO_SPACE_PART_DIV int = 3
	const DEDUCTION_OF_AVG float64 = 1 / 6.0
	//Space available for the actual graphs
	TotalBottomHeight := height - TOP_HEIGHT + 1
	//Average height for all Plots
	avgBottomItemHeight := TotalBottomHeight / COUNT_PLTS
	//The Rest of aboves division is added to graph1
	//Additional Space reserved for Load ans Soj.
	additionalSpace := 1

	if avgBottomItemHeight > MIN_AVG_BOT_HEIGHT {
		additionalSpace = util.FloorToInt(float64(avgBottomItemHeight) * (DEDUCTION_OF_AVG))
	}

	avgBottomItemHeight = util.MaxInt(avgBottomItemHeight, MIN_AVG_BOT_HEIGHT)
	if avgBottomItemHeight > 8 {
		avgBottomItemHeight -= additionalSpace
	}

	h1AdditionalSpace := 1
	h2AdditionalSpace := 0
	additionalSpace = TotalBottomHeight - avgBottomItemHeight*COUNT_PLTS
	//log.Printf("\nnewAvgBottomItemHeight=%d\nadditionalFreeSpace=%d\n", avgBottomItemHeight, additionalSpace)
	if additionalSpace == 2 {
		h2AdditionalSpace = 1
	} else if additionalSpace > 3 {
		additionalSpacePart := util.MaxInt(util.FloorToInt(float64(additionalSpace)/float64(PRIO_SPACE_PART_DIV)), 1)
		h1AdditionalSpace = additionalSpacePart*2 + additionalSpace%util.MaxInt(additionalSpacePart, 1)
		h2AdditionalSpace = additionalSpacePart

	}

	//log.Printf("\nh1additionalSpace=%d\nh2additionalSpace=%d\n", h1AdditionalSpace, h2AdditionalSpace)

	var H1 int = avgBottomItemHeight + h1AdditionalSpace
	var H2 int = avgBottomItemHeight + h2AdditionalSpace
	var H3 int = avgBottomItemHeight
	var H4 int = avgBottomItemHeight

	res := [4]int{H1, H2, H3, H4}
	hc := _heightConfig{
		topPart:        TOP_HEIGHT,
		bottomPart:     TotalBottomHeight,
		bottomSubParts: res,
	}
	//log.Println(hc.str())
	return hc
}

func generateLayout(ctx context.Context, t terminalapi.Terminal, man *flowdata.FlowManager, s *wids, layout int32, chans *coms.FlowChannels) ([]container.Option, error) {
	heights := plotHeights(t.Size().Y)
	topPart := grid.RowHeightFixed(heights.topPart,
		grid.ColWidthFixed(90,
			grid.Widget(s.I.InfoBox,
				container.Border(linestyle.Light),
				container.BorderTitle("Help"),
			)),
		grid.ColWidthFixed(30,
			grid.Widget(s.I.FlowDetails,
				container.Border(linestyle.Light),
				container.BorderTitle("Flow Details"),
			)),
		grid.ColWidthFixed(50,
			grid.Widget(s.I.FlowList,
				container.Border(linestyle.Light),
				container.BorderTitle("Flows"),
			)),
	)
	builder := grid.New()
	if layout == layoutCombined {
		builder.Add(
			topPart,
		)
	} else {
		if (s.M[layout] == detWids{}) {
			return nil, errors.New("cant show non existent Flow/Layout")
		}
		flowG := s.M[layout]
		builder.Add(
			topPart,
			grid.RowHeightFixed(heights.bottomSubParts[0],
				grid.Widget(flowG.LoadChart,
					container.Border(linestyle.Light),
					container.BorderTitle("Load [b/s]"),
				)),
			grid.RowHeightFixed(heights.bottomSubParts[1],
				grid.Widget(flowG.SojournChart,
					container.Border(linestyle.Light),
					container.BorderTitle("Sojourn time [ms]"),
				)),
			grid.RowHeightFixed(heights.bottomSubParts[2],
				grid.Widget(flowG.EcnChart,
					container.Border(linestyle.Light),
					container.BorderTitle("ECN [%]"),
				)),
			grid.RowHeightFixed(heights.bottomSubParts[3],
				grid.Widget(flowG.DroppChart,
					container.Border(linestyle.Light),
					container.BorderTitle("Dropped packages"),
				)),
		)
	}

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}
