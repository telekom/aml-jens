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

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

func NewDrpListTextBox(ctx context.Context, t terminalapi.Terminal, up *channel.DrpChannels) (*text.Text, error) {
	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}
	if err := wrapped.Write("Waiting for Flows", text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
		panic(err)
	}
	defaultOpts := text.WriteCellOpts()
	selectedOpts := text.WriteCellOpts(cell.Bold())

	go func() {
		for {
			select {
			case event := <-up.UpdateDrpList:
				wrapped.Reset()
				for i := 0; i < len(event.Drps); i++ {
					drp := event.Drps[i]
					opts := defaultOpts
					if i == event.Selected {
						opts = selectedOpts
						wrapped.Write(">", text.WriteCellOpts(cell.Blink()))
					} else {
						wrapped.Write(">")
					}
					err := wrapped.Write(fmt.Sprintf("  %s\n", drp.Name), opts)
					if err != nil {
						panic("Error while writing to flowlist")
					}
					if err != nil {
						panic("Error while writing to flowlist")
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return wrapped, nil
}
