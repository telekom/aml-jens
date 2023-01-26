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

package pipe

import (
	"context"
	"os"
	"time"

	coms "github.com/telekom/aml-jens/cmd/drshow/internal/data/channel"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/flowdata"
	"github.com/telekom/aml-jens/cmd/drshow/internal/data/widgets/shared"
	"github.com/telekom/aml-jens/cmd/drshow/internal/modes"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/logging"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

const rootID = "root"

var exitMsg *string = nil
var exitCode uint8 = 0
var fromError = func(e error) (uint8, string) {
	INFO.Println(e.Error())
	return 255, e.Error()
}
var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

func Run() (uint8, string) {

	manager := flowdata.NewFlowManager()
	reader, success := flowdata.ValidateDrpPlayerPipedIn()
	if !success {
		os.Exit(-1)
	}

	C := coms.NewFlowChannels()
	var t terminalapi.Terminal
	var err error
	t, err = tcell.New(tcell.ColorMode(terminalapi.ColorMode256))

	if err != nil {
		return fromError(err)
	}
	defer t.Close()
	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		return fromError(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	DEBUG.Printf("Terminal with size: %dx%d", t.Size().X, t.Size().Y)

	//All necessary components are defined. ManagerCallbacks
	//can now be defined
	manager.SetHandler(flowdata.FlowManagerEventHanderlerT{
		OnEndOfDrpplayer: func() {
			upd := shared.NewStrTextWriteOps("DataRatePlayer Ended!")
			upd.Opts = append(upd.Opts, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			C.UpdateText <- upd
		},
		NewFlow: func(man *flowdata.FlowManager) error {
			C.UpdateFlowList <- man

			return nil
		},
		NewFlowSelected: func(man *flowdata.FlowManager) error {
			C.UpdateFlowList <- man
			flow, err := man.GetSelectedFlow()
			if err != nil {
				WARN.Printf("Can get Selected Flow from Manger: %+v", err)
			}
			C.UpdateFlowDetails <- flow
			return nil
		},
		NewDataInFlow: func(flow *flowdata.FlowT) error {
			C.UpdateDataGraphs <- flow
			return nil
		},

		OnFirstFlowAdded: func(man *flowdata.FlowManager) {
			flow, _ := man.GetSelectedFlow()
			C.RedrawUIWithLayout <- flow.FlowId
		},
		OnFailureOrEnd: func(code uint8, text string) {
			t.Flush()
			cancel()
			exitMsg = &text
		},
	})

	manager.MainSTDinReadLoop(reader)

	//Common instructions for both increment and decrement
	cbFlowListAfter := func() error {
		flow, err := manager.GetSelectedFlow()
		if err != nil {
			return nil
		}
		C.RedrawUIWithLayout <- flow.FlowId
		return nil
	}
	cbFlowListInc := func() error {
		err := manager.SelectNextFlowIndex()
		if err == nil {
			return cbFlowListAfter()
		}
		return nil
	}
	cbFlowListDec := func() error {
		err := manager.SelectPreviousFlowIndex()
		if err == nil {
			return cbFlowListAfter()
		}
		return nil
	}

	//Create CustomUI Object
	ui, err := NewUI(ctx, t, c, &uiSettings{
		cb_flow_listInc: cbFlowListInc,
		cb_flow_listDec: cbFlowListDec,
	}, C, manager)
	if err != nil {
		cancel()
		return fromError(err)
	}

	if err := setLayout(ctx, t, c, manager, ui, -1, C); err != nil {
		manager.ExitApplicationErr(err.Error())
	}
	go func() {
		for {
			<-time.After(500 * time.Millisecond)
			var flow *flowdata.FlowT
			var err error
			if flow, err = manager.GetSelectedFlow(); err != nil {
				continue
			}
			setLayout(ctx, t, c, manager, ui, flow.FlowId, C)
		}
	}()
	//Redraw Loop. Runs in background.
	//Checks appropriate Channel, if a new flow should
	//be displayed.
	go func() {
		for {
			select {
			case lt := <-C.RedrawUIWithLayout:
				if lt == layoutCombined {
					C.UpdateFlowDetails <- nil
				} else {
					f, err := manager.GetSelectedFlow()
					if err != nil {
						manager.ExitApplicationErr(err.Error())
					}
					C.UpdateFlowDetails <- f
				}
				setLayout(ctx, t, c, manager, ui, lt, C)

			}
		}
	}()
	keyboardListener := func(k *terminalapi.Keyboard) {
		if modes.IsExitKey(k.Key) {
			manager.ExitApplicationOK("Goodbye")
			return
		}
		switch k.Key {
		case keyboard.KeyArrowUp:
			if cbFlowListDec() != nil {
				manager.ExitApplicationErr("Cant decrement")
			}
		case keyboard.KeyArrowDown:
			if cbFlowListInc() != nil {
				manager.ExitApplicationErr("Cant increment")
			}
		case keyboard.KeySpace:
			flow, err := manager.GetSelectedFlow()
			if err == nil {
				C.RedrawUIWithLayout <- flow.FlowId
			}
		case 'e', 'E':
			flow, err := manager.GetSelectedFlow()
			if err == nil {
				go func() {
					err := flow.ExportToFile(config.ShowCfg().ExportPathPrefix)
					if err != nil {
						INFO.Printf("Error while exporting [%d]: %s\n", manager.GetSelectedFlowIndex(), err.Error())
					}
				}()
			}
		default:

		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(keyboardListener), termdash.RedrawInterval(50*time.Millisecond)); err != nil {
		manager.ExitApplicationErr(err.Error())
	}

	//Exit
	os.Stdin.Close()
	return exitCode, *exitMsg
}

// Set (, Generate) and Update the Terminal with a 'new' UI
func setLayout(ctx context.Context, t terminalapi.Terminal, c *container.Container, man *flowdata.FlowManager, w *wids, lt int32, chans *coms.FlowChannels) error {
	gridOpts, err := generateLayout(ctx, t, man, w, lt, chans)
	if err != nil {
		return err
	}

	return c.Update(rootID, gridOpts...)
}
