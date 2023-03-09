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

package folder

import (
	"context"
	"errors"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	drpWidgets "github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/drp"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/shared"
	"github.com/telekom/aml-jens/pkg/drp"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

type wids struct {
	drpCharts []*linechart.LineChart
	textField *text.Text
	list      *text.Text
	infoBox   *text.Text
}

func NewUI(ctx context.Context, t terminalapi.Terminal, c *container.Container, updater *channel.DrpChannels, patterns *drp.DrpListT) (*wids, error) {
	txt, err := drpWidgets.NewDrpDetailsTextBox(ctx, t, updater)
	if err != nil {
		return nil, err
	}
	if patterns.Selected == -1 {
		return nil, errors.New("no pattern(s) found")
	}
	drpGraphs := []*linechart.LineChart{}
	for i := 0; i < len(patterns.Drps); i++ {
		drp, err := drpWidgets.NewStaticPatternLineChart(ctx, &patterns.Drps[i])

		if err != nil {
			return nil, err
		}
		drpGraphs = append(drpGraphs, drp)
	}

	list, err := drpWidgets.NewDrpListTextBox(ctx, t, updater)
	if err != nil {
		return nil, err
	}
	info, err := shared.NewStaticTextBoxQuitMessage(ctx, t,
		[]shared.StrWithTextOpts{
			shared.NewStrTextWriteOps("Nav: <ArrowUp> & <ArrowDown>"),
		})
	if err != nil {
		return nil, err
	}
	return &wids{
		drpCharts: drpGraphs,
		textField: txt,
		list:      list,
		infoBox:   info,
	}, nil
}
func (s *wids) Layout(patterns *drp.DrpListT) ([]container.Option, error) {
	builder := grid.New()
	builder.Add(
		grid.ColWidthFixed(30,
			grid.RowHeightFixedWithOpts(4, []container.Option{
				container.Border(linestyle.Light),
				container.BorderTitle("Information"),
				container.BorderTitleAlignCenter(),
			},
				grid.Widget(s.infoBox),
			),
			grid.RowHeightFixedWithOpts(8, []container.Option{
				container.Border(linestyle.Light),
				container.BorderTitle("Pattern Information"),
				container.BorderTitleAlignCenter(),
			},
				grid.Widget(s.textField),
			),

			grid.RowHeightFixedWithOpts(10, []container.Option{
				container.Border(linestyle.Light),
				container.BorderTitle("Pattern List"),
				container.BorderTitleAlignCenter(),
			},
				grid.Widget(s.list),
			),
		),
		grid.RowHeightPercWithOpts(50, []container.Option{
			container.Border(linestyle.Light),
			container.BorderTitle("Pattern: " + patterns.GetSelected().Name),
			container.BorderTitleAlignCenter(),
		}, grid.Widget(s.drpCharts[patterns.Selected])),
	)
	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}
