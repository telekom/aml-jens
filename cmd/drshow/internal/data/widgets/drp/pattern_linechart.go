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

package drp

import (
	"context"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/drpdata"
	"github.com/telekom/aml-jens/internal/util"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
)

func NewPatternLineChart(ctx context.Context, updater *channel.DrpChannels) (*linechart.LineChart, error) {
	chart, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.YAxisFormattedValues(util.FormatLabelISOKilo),
		linechart.YAxisAdaptive(),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case drp := <-updater.UpdateDrpPattern:
				err = chart.Series(drp.Name, *drp.GetData(), linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))))
				if err != nil {
					panic(err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return chart, nil
}
func NewStaticPatternLineChart(ctx context.Context, drp *drpdata.DrpT) (*linechart.LineChart, error) {
	chart, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.YAxisFormattedValues(util.FormatLabelISOKilo),
		linechart.YAxisAdaptive(),
	)
	if err != nil {
		return nil, err
	}
	err = chart.Series(drp.Name, *drp.GetData(), linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))))
	if err != nil {
		panic(err)
	}
	return chart, nil
}
