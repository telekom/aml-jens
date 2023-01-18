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
	"time"

	"github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/cmd/drshow/internal/modes"
	"github.com/telekom/aml-jens/pkg/drp"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

const rootID = "root"

var ErrorOrNil error = nil

func Run(path string) error {
	drpList, err := drp.NewDrpListFromFolder(path)
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
	updateChannel := channel.NewStaticDrpChannels()
	ui, err := NewUI(ctx, t, c, updateChannel, drpList)
	if err != nil {
		cancel()
		return err
	}
	gridOpts, err := ui.Layout(drpList) // equivalent to contLayout(w)
	if err != nil {
		cancel()
		return err
	}

	if err := c.Update(rootID, gridOpts...); err != nil {
		cancel()
		return err
	}
	navigator := func(k *terminalapi.Keyboard) {
		if modes.IsExitKey(k.Key) {
			cancel()
		}
		defer func() {
			updateChannel.UpdateDrpGlobally(drpList.Drps[drpList.Selected])
			updateChannel.UpdateDrpList <- *drpList
			gridOpts, err := ui.Layout(drpList)
			if err != nil {
				ErrorOrNil = err
				cancel()
			}
			if err := c.Update(rootID, gridOpts...); err != nil {
				ErrorOrNil = err
				cancel()
			}
		}()
		if k.Key == keyboard.KeyArrowUp {
			newI := drpList.Selected - 1
			if newI < 0 {
				return
			}
			drpList.Selected = newI
			return
		}
		if k.Key == keyboard.KeyArrowDown {
			newI := drpList.Selected + 1
			if newI >= len(drpList.Drps) {
				return
			}
			drpList.Selected = newI
		}
	}
	updateChannel.UpdateDrpList <- *drpList

	updateChannel.UpdateDrpGlobally(drpList.Drps[drpList.Selected])
	if err := termdash.Run(ctx, t, c,
		termdash.RedrawInterval(50*time.Millisecond),
		termdash.KeyboardSubscriber(navigator),
	); err != nil {
		return err
	}
	return ErrorOrNil
}
