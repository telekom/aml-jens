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

package flowdata

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/logging"
)

const (
	i_timestamp = iota
	i_sojourntime
	i_load
	i_capacity
	i_ecn
	i_drop
	i_netw
)
const (
	in_src = iota
	in_dst
)

var CFG = config.ShowCfg()
var DEBUG, INFO, FATAL = logging.GetLogger()

type FlowManager struct {
	Mutex             sync.Mutex
	Flows             []*FlowT
	Handler           FlowManagerEventHanderlerT
	selectedFlowIndex int
	version           uint64
}

func NewFlowManager() *FlowManager {
	return &FlowManager{
		Flows:             make([]*FlowT, 0, 256),
		selectedFlowIndex: 0,
		version:           1,
	}
}
func (man *FlowManager) GetSelectedFlow() (*FlowT, error) {
	if man.selectedFlowIndex == -1 || man.selectedFlowIndex >= len(man.Flows) {
		return nil, fmt.Errorf("cant get flow with index=%d", man.selectedFlowIndex)
	}
	return man.Flows[man.selectedFlowIndex], nil
}
func (man *FlowManager) GetSelectedFlowIndex() int {
	return man.selectedFlowIndex
}

func (man *FlowManager) SetHandler(handler FlowManagerEventHanderlerT) {
	man.Handler = handler
}
func (man *FlowManager) ExitApplicationOK(message string) {
	man.Handler.OnFailureOrEnd(0, message)
}
func (man *FlowManager) ExitApplicationErr(errorMsg string) {
	man.Handler.OnFailureOrEnd(255, errorMsg)
}

func (man *FlowManager) Version() uint64 {
	return man.version
}

// Selects next Eligible Flow
func (man *FlowManager) SelectNextFlowIndex() error {
	man.Mutex.Lock()
	defer man.Mutex.Unlock()
	relIndexOfSelected, absoluteIndexes, _ := man.EligibleFLows()
	if len(absoluteIndexes) <= 1 {
		return errors.New("no Flows are eligable/ existent")
	}
	if relIndexOfSelected.MatchType == nextMatch {
		man.selectedFlowIndex = absoluteIndexes[relIndexOfSelected.Index]
	} else {
		if relIndexOfSelected.Index == len(absoluteIndexes)-1 {
			return errors.New("cant increment SelectedFlowIndex. It is the maxvalue")
		}
		man.selectedFlowIndex = absoluteIndexes[relIndexOfSelected.Index+1]
	}

	if err := man.Handler.NewFlowSelected(man); err != nil {
		man.Handler.OnFailureOrEnd(255, err.Error())
	}
	return nil
}

// Selects previous Eligible Flow
func (man *FlowManager) SelectPreviousFlowIndex() error {
	man.Mutex.Lock()
	defer man.Mutex.Unlock()
	relIndexOfSelected, absoluteIndexes, _ := man.EligibleFLows()
	if relIndexOfSelected.MatchType == nextMatch {
		relIndexOfSelected.Index--
	}
	if len(absoluteIndexes) <= 1 || relIndexOfSelected.Index == 0 {
		return errors.New("cant decrement SelectedFlowIndex. Either no other flows or at the beginning")
	}
	man.selectedFlowIndex = absoluteIndexes[relIndexOfSelected.Index-1]
	if err := man.Handler.NewFlowSelected(man); err != nil {

		man.Handler.OnFailureOrEnd(255, err.Error())
	}
	return nil
}
func (man *FlowManager) GetCopyById(id int32) *FlowDatapointsT {
	man.Mutex.Lock()
	defer man.Mutex.Unlock()
	var flow *FlowT
	for _, v := range man.Flows {
		if v.FlowId == id {
			flow = v
		}
	}

	if flow == nil {
		return nil
	}
	return &flow.D
}
func (man *FlowManager) get(f *FlowT) *FlowT {
	man.Mutex.Lock()
	defer man.Mutex.Unlock()
	for _, v := range man.Flows {
		if v.Equals(f) {
			return v
		}
	}
	return nil
}
func (man *FlowManager) GetAndAppendTo(f *FlowT, ts float64, sj float64, ld float64, ec float64, dr float64, cap float64) {
	v := man.get(f)
	v.D.Append(ts, sj, ld, ec, dr, cap)

	if err := man.Handler.NewDataInFlow(v); err != nil {
		man.ExitApplicationErr(err.Error())
	}
}

// Add a new Flow to manager
func (man *FlowManager) Append(f *FlowT) {
	man.Mutex.Lock()
	f.FlowId = int32(len(man.Flows) + 1)
	man.Flows = append(man.Flows, f)
	man.version++
	man.Mutex.Unlock()
	err := man.Handler.NewFlow(man)
	if err != nil {
		man.ExitApplicationErr(err.Error())
	}
	if len(man.Flows) == 1 {
		if man.selectedFlowIndex != 0 {
			man.ExitApplicationErr("unknown selected flow")
		}
		man.Handler.NewFlowSelected(man)
		man.Handler.OnFirstFlowAdded(man)
	}
}
func (man *FlowManager) FmtString() string {
	ret := "FlowManger{\n"
	for i := 0; i < len(man.Flows); i++ {
		ret += fmt.Sprintf("  %d:%d\n", man.Flows[i].FlowId, man.Flows[i].D.Length())
	}
	return ret + "}\n"
}
func (man *FlowManager) Contains(f *FlowT) bool {
	man.Mutex.Lock()
	defer man.Mutex.Unlock()
	for _, v := range man.Flows {
		if v.Equals(f) {
			return true
		}
	}
	return false
}
func (man *FlowManager) MainSTDinReadLoop(r *bufio.Reader) {

	go func() {
		for {
			line, err := r.ReadString('\n')
			line = strings.TrimSuffix(line, "\n")
			if err != nil {
				if err == io.EOF {
					log.Println("EOF found while reading from STDIN")
					man.Handler.OnEndOfDrpplayer()
					return
				} else {
					FATAL.Println(err.Error())
					man.ExitApplicationErr("Error while Reading: " + err.Error())
				}

			}
			if !man.ReadFromLine(line) { //"End of drplay"
				INFO.Println("Something was skipped")
				//man.Handler.OnEndOfDrpplayer()
			}
		}
	}()
}
func setupReader() *bufio.Reader {
	reader := bufio.NewReader(os.Stdin)
	return reader
}
func ValidateDrpPlayerPipedIn() (*bufio.Reader, bool) {
	r := setupReader()
	input_skipped := make(chan *[]string)
	defer close(input_skipped)
	go skipToMeasures(r, input_skipped)
	select {
	case lines := <-input_skipped:
		if lines != nil {
			for _, v := range *lines {
				log.Fatal(v)
			}
		}
		return r, lines == nil
	case <-time.After(time.Duration(CFG.WaitTimeForSamplesInSec) * time.Second):
		log.Printf("No input recieved from stdin after %d seconds. This can be increased in the config file (%s).", CFG.WaitTimeForSamplesInSec, paths.RELEASE_CFG_PATH()+assets.CFG_FILE_NAME)
		log.New(os.Stderr, "", 0).Fatalf("No input recieved from stdin after %d seconds.", CFG.WaitTimeForSamplesInSec)
		return nil, false
	}
}
func skipToMeasures(r *bufio.Reader, c chan *[]string) {
	lines := make([]string, 256)
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSuffix(line, "\n")
		lines = append(lines, line)
		//log.Printf("(%s)\n", line)
		if err == io.EOF {
			log.Print("EOF recieved, maybe nothing was piped in?")
			c <- &lines
			return
		}
		if isMagicHeader(strings.Fields(line)) {
			log.Println("Magic Header found. Listening for samples now.")
			c <- nil
			break
		}
	}
}

func isMagicHeader(arr []string) bool {
	if len(arr) != len(assets.CONST_HEADING) {
		return false
	}
	for i, v := range assets.CONST_HEADING {
		if v != arr[i] {
			return false
		}
	}
	return true
}

func (manager *FlowManager) ReadFromLine(line string) bool {
	if strings.HasPrefix(line, "http") {
		return true
	}
	splitData := strings.Split(line, " ")
	parseFloat := func(ln string) float64 {
		i, err := strconv.ParseFloat(ln, 64)
		if err != nil {
			INFO.Printf("Invalid Line format. Not a Number: %+v", err)
			return 0
		}
		return i
	}
	if len(splitData) != len(assets.CONST_HEADING) {
		INFO.Printf("Invalid Line format. Wrong length (%d):'%s'\n", len(splitData), line)
		return false
	}
	net_data := strings.Split(splitData[i_netw], "-")
	if len(net_data) != 2 {
		INFO.Fatalf("Did not find correct flow format: %s", splitData[i_netw])
		return false
	}
	f := NewFlow(
		net_data[in_src],
		net_data[in_dst],
	)
	//manager.Mutex.Lock()
	if !manager.Contains(f) {
		manager.Append(f)
	}

	go manager.GetAndAppendTo(f,
		parseFloat(splitData[i_timestamp]),
		parseFloat(splitData[i_sojourntime]),
		parseFloat(splitData[i_load]),
		parseFloat(splitData[i_ecn]),
		parseFloat(splitData[i_drop]),
		parseFloat(splitData[i_capacity]),
	)
	//manager.Mutex.Unlock()
	return true
}
