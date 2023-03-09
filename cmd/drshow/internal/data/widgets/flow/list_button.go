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

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/button"
)

type FlowListButtonSettings struct {
	Title    string
	ColorNbr uint8
}

func FlowLBtnUP() FlowListButtonSettings {
	return FlowListButtonSettings{
		Title:    "  Next  ",
		ColorNbr: 4,
	}
}
func FlowLBtnDown() FlowListButtonSettings {
	return FlowListButtonSettings{
		Title:    "Previous",
		ColorNbr: 4,
	}
}
func FlowLBtnLayoutCombined() FlowListButtonSettings {
	return FlowListButtonSettings{
		Title:    "All",
		ColorNbr: 200,
	}
}
func FlowLBtnLayoutSingle() FlowListButtonSettings {
	return FlowListButtonSettings{
		Title:    "Single",
		ColorNbr: 200,
	}
}

type BtnFlowListType uint8

const (
	DOWN = iota
	UP
	LAYOUT_COMBINED
	LAYOUT_SINGLE
)

func NewFlowListButton(ctx context.Context, t terminalapi.Terminal, typ FlowListButtonSettings, cb button.CallbackFn) (*button.Button, error) {
	return button.New(
		typ.Title,
		cb,
		button.DisableShadow(),
		button.Height(1),
		button.TextColor(cell.ColorNumber(int(typ.ColorNbr))),
	)
}
