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
	"fmt"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/util"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

func NewDrpDetailsTextBox(ctx context.Context, t terminalapi.Terminal, chans *channel.DrpChannels) (*text.Text, error) {
	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		return nil, err
	}
	if err := wrapped.Write("No Flow selected", text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case drp := <-chans.UpdateDrpDetails:
				wrapped.Reset()
				err := wrapped.Write(fmt.Sprintf("Pattern: %s\nSamples: %d\nMinimum: %s\nMaximum: %s\nAverage: %s\nPath:    %s", drp.Name, drp.SampleCount(), util.FormatLabelISOKilo(drp.Min), util.FormatLabelISOKilo(drp.Max), util.FormatLabelISOKilo(drp.Avg), drp.GetOrigin()))
				if err != nil {
					WARN.Println(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return wrapped, nil
}
