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
	"time"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/acquisition"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/cmd/drshow/internal/modes"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

const rootID = "root"

var ErrorOrNil error = nil

func Run(path string) error {
	drp, err := acquisition.NewDrpFromFile(path)
	if err != nil {
		return err
	}
	var t terminalapi.Terminal
	t, err = tcell.New(tcell.ColorMode(terminalapi.ColorMode256))

	if err != nil {
		return err
	}
	defer t.Close()
	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	updateChannel := channel.NewDrpChannels()
	ui, err := NewUI(ctx, t, c, updateChannel)
	if err != nil {
		cancel()
		return err

	}
	gridOpts, err := ui.Layout() // equivalent to contLayout(w)
	if err != nil {
		cancel()
		return err
	}

	if err := c.Update(rootID, gridOpts...); err != nil {
		cancel()
		return err
	}

	quitter := func(k *terminalapi.Keyboard) {
		if modes.IsExitKey(k.Key) {
			cancel()
		}
	}
	updateChannel.UpdateDrpGlobally(*drp)
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(50*time.Millisecond)); err != nil {
		return err
	}
	return ErrorOrNil
}
