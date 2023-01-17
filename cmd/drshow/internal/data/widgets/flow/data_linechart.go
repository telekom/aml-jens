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

	data "github.com/telekom/aml-jens/cmd/drshow/internal/data/flowdata"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/util"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
)

var CFG = config.ShowCfg()
var label_color = cell.ColorWhite

func NewFlowDataLineChartPerc(ctx context.Context, dataIn <-chan data.DisplayDataT, flowId int32) (*linechart.LineChart, error) {
	percentFormatter := func(value float64) string {
		return fmt.Sprintf(" %3.0f", value)
	}
	var chart *linechart.LineChart
	var err error
	if CFG.PlotScrollMode == config.Scrolling {
		chart, err = linechart.New(
			linechart.YAxisCustomScale(0, 101),
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(percentFormatter),
			linechart.XAxisUnscaled(),
		)
	} else {
		chart, err = linechart.New(
			linechart.YAxisCustomScale(0, 101),
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(percentFormatter),
		)
	}

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case newData := <-dataIn:
				if flowId < 0 || newData.FlowId == flowId {
					err = chart.Series(newData.Name, *newData.Data,
						linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(int(newData.Color)))))
					if err != nil {
						panic(err)
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return chart, nil
}
func NewFlowDataLineChart(ctx context.Context, dataIn <-chan data.DisplayDataT, flowId int32) (*linechart.LineChart, error) {
	var chart *linechart.LineChart
	var err error
	if CFG.PlotScrollMode == config.Scrolling {
		chart, err = linechart.New(
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISOKilo),
			linechart.XAxisUnscaled(),
		)
	} else {
		chart, err = linechart.New(
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISOKilo),
		)
	}

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case newData := <-dataIn:
				if flowId < 0 || newData.FlowId == flowId {
					err = chart.Series(newData.Name, *newData.Data,
						linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(int(newData.Color)))))
					if err != nil {
						panic("newDataError: " + err.Error())
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return chart, nil
}
func NewFlowDataLineChartCombinationLoadCapacity(ctx context.Context, dataIn <-chan data.DisplayDataDualT, flowId int32) (*linechart.LineChart, error) {
	var chart *linechart.LineChart
	var err error
	if CFG.PlotScrollMode == config.Scrolling {
		chart, err = linechart.New(
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISOKilo),
			linechart.XAxisUnscaled(),
		)
	} else {
		chart, err = linechart.New(
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISOKilo),
		)
	}

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case newData := <-dataIn:
				if flowId < 0 || newData.FlowId == flowId {
					err = chart.Series(newData.Name, *newData.Data,
						linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(int(newData.Color)))))
					if err != nil {
						panic("newDataError: " + err.Error())
					}
					err = chart.Series("Linkcapacity", *newData.DataExtra,
						linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(15))))
					if err != nil {
						panic("newDataError: " + err.Error())
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return chart, nil
}

func NewFlowDataLineChartAbs(ctx context.Context, dataIn <-chan data.DisplayDataT, flowId int32) (*linechart.LineChart, error) {
	var chart *linechart.LineChart
	var err error
	if CFG.PlotScrollMode == config.Scrolling {
		chart, err = linechart.New(
			linechart.YAxisCustomScale(0, 1),
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISO),
			linechart.XAxisUnscaled(),
		)
	} else {
		chart, err = linechart.New(
			linechart.YAxisCustomScale(0, 1),
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISO),
		)
	}

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case newData := <-dataIn:
				if flowId < 0 || newData.FlowId == flowId {
					err = chart.Series(newData.Name, *newData.Data,
						linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(int(newData.Color)))))
					if err != nil {
						panic("newDataError: " + err.Error())
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return chart, nil
}

func NewFlowDataLineChartMs(ctx context.Context, dataIn <-chan data.DisplayDataT, flowId int32) (*linechart.LineChart, error) {
	var chart *linechart.LineChart
	var err error
	if CFG.PlotScrollMode == config.Scrolling {
		chart, err = linechart.New(
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISO),
			linechart.XAxisUnscaled(),
		)
	} else {
		chart, err = linechart.New(
			linechart.AxesCellOpts(cell.FgColor(cell.ColorWhite)),
			linechart.YLabelCellOpts(cell.FgColor(label_color)),
			linechart.XLabelCellOpts(cell.FgColor(label_color)),
			linechart.YAxisFormattedValues(util.FormatLabelISO),
		)
	}

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case newData := <-dataIn:
				if flowId < 0 || newData.FlowId == flowId {
					err = chart.Series(newData.Name, *newData.Data,
						linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(int(newData.Color)))))
					if err != nil {
						panic(err)
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return chart, nil
}
