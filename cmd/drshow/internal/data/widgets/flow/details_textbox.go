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
	"log"
	"time"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/flowdata"
	"github.com/telekom/aml-jens/internal/util"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

func NewFlowDetailsTextBox(ctx context.Context, t terminalapi.Terminal, chans *channel.FlowChannels) (*text.Text, error) {
	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}
	if err := wrapped.Write("Make a flow-selection\n(SingleMode)", text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))); err != nil {
		panic(err)
	}
	updateText := func(flow *flowdata.FlowT) {
		wrapped.Reset()
		if flow == nil {
			log.Fatal("chans.UpdateFlowDetails received nil")
		}
		var timePassed string
		if len(*flow.D.TimeStamp) == 0 {
			timePassed = "0"
		} else {

			timePassed = time.Since(time.UnixMilli(int64((*flow.D.TimeStamp)[len(*flow.D.TimeStamp)-1]))).Round(time.Second).String()
		}
		txt := fmt.Sprintf("ID:   %04d\nSrc:%s\nDst:%s\nCount:%s\nLastSample:%s ago",
			flow.FlowId,
			flow.Src.Str(),
			flow.Dst.Str(),
			util.FormatLabelISO(float64(flow.D.Length())),
			timePassed,
		)

		err := wrapped.Write(txt)
		if err != nil {
			log.Fatal(err)
			panic(err)

		}
	}
	go func() {
		var lastSelectedFlow *flowdata.FlowT = nil
		for {
			select {
			case <-time.After(time.Second * 5):
				if lastSelectedFlow != nil {
					updateText(lastSelectedFlow)
				}
			case flow := <-chans.UpdateFlowDetails:
				lastSelectedFlow = flow
				updateText(flow)
			case <-ctx.Done():
				return
			}
		}
	}()
	return wrapped, nil
}
