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

package flow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/flowdata"
	"github.com/telekom/aml-jens/internal/util"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

func NewFlowListTextBox(ctx context.Context, t terminalapi.Terminal, updateFlows <-chan *flowdata.FlowManager) (*text.Text, error) {
	wrapped, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := wrapped.Write("Waiting for Flows", text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
		panic(err)
	}

	const ROW_COUNT int = 6
	go func() {
		var lastUpdate = time.Now()
		for {
			//+ writeBlinkingOptIf(bSelectedFlow, "■", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(int(flow.ColorI)))))
			select {
			case man := <-updateFlows:
				if time.Now().Sub(lastUpdate) < time.Second {
					continue
				}
				wrapped.Reset()
				_, absoluteIndexes, eligableFlows := man.EligibleFLows()
				col_count := len(absoluteIndexes) / ROW_COUNT
				if len(absoluteIndexes)%ROW_COUNT != 0 {
					col_count++
				}
				//panic(fmt.Sprintf("RowCount=%d, ColCount=%f"))

				colC, rowC, dataTablePtr := convertFlowsTo2dArray(eligableFlows, ROW_COUNT)
				if dataTablePtr == nil {
					break
				}
				print2dArrayOfFlows(wrapped, dataTablePtr, rowC, colC, man)
			case <-ctx.Done():
				return
			}
		}
	}()
	return wrapped, nil
}
func writeHighlightedIf(wrapped *text.Text, b bool, t string) error {
	if b {

		return wrapped.Write(t, text.WriteCellOpts(cell.BgColor(cell.ColorBlack), cell.Bold()))
	} else {
		return wrapped.Write(t)
	}
}
func writeHiglightedColoredIf(wrapped *text.Text, b bool, t string, color uint8) error {
	if b {

		return wrapped.Write(t, text.WriteCellOpts(cell.BgColor(cell.ColorBlack), cell.Bold(), cell.FgColor(cell.ColorNumber(int(color)))))
	} else {
		return wrapped.Write(t)
	}
}
func print2dArrayOfFlows(wrapped *text.Text, dataTable *[][]*flowdata.FlowT, rowC int, colC int, m *flowdata.FlowManager) {
	selectedFlow, _ := m.GetSelectedFlow()
	defer wrapped.Write(strings.Repeat("       ⤻    ", colC-1))

	writeLine := func(lineNbr int, t *[][]*flowdata.FlowT) {
		defer wrapped.Write("\n")
		for i := 0; i < len(*dataTable); i++ {
			flow := (*t)[i][lineNbr]
			if flow == nil {
				continue
			}
			isSelectedFlow := flow.Equals(selectedFlow)
			if err := writeHighlightedIf(
				wrapped,
				isSelectedFlow,
				fmt.Sprintf("%04d", flow.FlowId),
			); err != nil {
				panic(err)
			}
			writeHiglightedColoredIf(wrapped, isSelectedFlow,
				"(",
				flow.Color(),
			)
			writeHighlightedIf(wrapped, isSelectedFlow,
				fmt.Sprintf("% 5s", util.FormatLabelISOShorter(float64(flow.D.Length()))))
			writeHiglightedColoredIf(wrapped, isSelectedFlow,
				")",
				flow.Color(),
			)
			if err := wrapped.Write("▍",
				text.WriteCellOpts(cell.FgColor(cell.ColorNumber(int(flow.Color()))))); err != nil {
				panic(err)
			}
			/*if err := wrapped.Write("⎸"); err != nil {
				panic(err)
			}*/
		}
	} //tc qdisc delete dev lo parent 1:1

	for line := 0; line < rowC; line++ {
		writeLine(line, dataTable)
	}
}
func convertFlowsTo2dArray(eligibleFlows []*flowdata.FlowT, rowC int) (int, int, *[][]*flowdata.FlowT) {
	if len(eligibleFlows) == 0 {
		return 0, 0, nil
	}
	colC := len(eligibleFlows) / rowC
	if len(eligibleFlows)%rowC != 0 {
		colC++
	}
	t := make([][]*flowdata.FlowT, colC)
	for i, _ := range t {
		t[i] = make([]*flowdata.FlowT, rowC)
	}

	for i := 0; i < len(eligibleFlows); i++ {
		t[i/rowC][i%rowC] = eligibleFlows[i]
	}
	return colC, rowC, &t
}
