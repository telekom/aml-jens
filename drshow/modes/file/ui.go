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

package file

import (
	"context"

	"jens/drshow/data/channel"
	drpWidgets "jens/drshow/data/widgets/drp"
	"jens/drshow/data/widgets/shared"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

type wids struct {
	drpChart  *linechart.LineChart
	textField *text.Text
	info      *text.Text
}

func NewUI(ctx context.Context, t terminalapi.Terminal, c *container.Container, updater *channel.DrpChannels) (*wids, error) {
	txt, err := drpWidgets.NewDrpDetailsTextBox(ctx, t, updater)
	if err != nil {
		return nil, err
	}
	drp, err := drpWidgets.NewPatternLineChart(ctx, updater)
	if err != nil {
		return nil, err
	}
	info, err := shared.NewStaticTextBoxQuitMessage(ctx, t,
		[]shared.StrWithTextOpts{})
	if err != nil {
		return nil, err
	}
	return &wids{
		drpChart:  drp,
		textField: txt,
		info:      info,
	}, nil
}
func (s *wids) Layout() ([]container.Option, error) {
	builder := grid.New()
	builder.Add(
		grid.ColWidthFixed(30,
			grid.RowHeightFixedWithOpts(3, []container.Option{
				container.Border(linestyle.Light),
				container.BorderTitle("Information"),
				container.BorderTitleAlignCenter(),
			},
				grid.Widget(s.info),
			),
			grid.RowHeightFixedWithOpts(10, []container.Option{
				container.Border(linestyle.Light),
				container.BorderTitle("Pattern Information"),
				container.BorderTitleAlignCenter(),
			},
				grid.Widget(s.textField),
			),
		),
		grid.ColWidthFixedWithOpts(50, []container.Option{
			container.Border(linestyle.Light),
			container.BorderTitle("Pattern"),
			container.BorderTitleAlignCenter(),
		}, grid.Widget(s.drpChart)),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}
