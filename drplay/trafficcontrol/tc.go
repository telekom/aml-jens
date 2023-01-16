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

package trafficcontrol

import (
	"encoding/binary"
	"fmt"
	"jens/drcommon/assets"
	"jens/drcommon/commands"
	"jens/drcommon/persistence/datatypes"

	"jens/drcommon/errortypes"
	"jens/drcommon/logging"
	"os"
	"sync"
	"time"
)

var DEBUG, INFO, FATAL = logging.GetLogger()

const CTRL_FILE = "/sys/kernel/debug/sch_janz/0001:v1"

type TrafficControlStartParams struct {
	Datarate     uint32
	QueueSize    int
	AddonLatency int
	Markfree     int
	Markfull     int
}
type NftStartParams struct {
	L4sPremarking bool
	SignalStart   bool
}

func (p TrafficControlStartParams) validate() error {
	if p.Datarate < 100 {
		return errortypes.NewUserInputError("Datarate should not be below 100")
	}
	if p.AddonLatency >= 10000 {
		return errortypes.NewUserInputError("AddonLatency should not be above 10000ms")
	}
	if p.Markfree > p.Markfull {
		return errortypes.NewUserInputError("Markfree should not be greater Markfull")
	}
	if p.Markfree > 0xffff {
		return errortypes.NewUserInputError("Markfree possibly corrupted")
	}
	if p.Markfull > 0xffff {
		return errortypes.NewUserInputError("Markfull possibly corrupted")
	}
	return nil
}

func (p TrafficControlStartParams) asArgs() []string {
	args := make([]string, 0, 5)
	args = append(args, "rate", fmt.Sprintf("%dkbit", p.Datarate))

	if p.QueueSize > 0 {
		args = append(args, "limit", fmt.Sprintf("%d", p.QueueSize))
	}
	if p.AddonLatency > 0 {
		args = append(args, "extralatency", fmt.Sprintf("%dms", p.AddonLatency))
	}
	if p.Markfree > 0 {
		args = append(args, "markfree", fmt.Sprintf("%dms", p.Markfree))
	}
	if p.Markfull > 0 {
		args = append(args, "markfull", fmt.Sprintf("%dms", p.Markfull))
	}
	return args
}

type TrafficControl struct {
	dev               string
	current_data_rate float64
	control_file      *os.File
	nft               NftStartParams
}

func NewTrafficControl(dev string) *TrafficControl {
	tc := &TrafficControl{
		dev: dev,
	}
	return tc
}

func (tc *TrafficControl) Init(params TrafficControlStartParams, nft NftStartParams) error {
	if nft.L4sPremarking {
		ResetECTMarking(assets.NFT_TABLE_PREMARK)
		CreateNftRuleECT(tc.dev, assets.NFT_TABLE_PREMARK, assets.NFT_CHAIN_FORWARD, assets.NFT_CHAIN_OUTPUT, "ect1", "0")
	}

	if err := params.validate(); err != nil {
		return err
	}
	if err := tc.Reset(); err != nil {
		DEBUG.Printf("TcStart>TcReset: %v", err)
	}
	args := []string{"qdisc", "add", "dev", tc.dev, "root", "handle", "1:", "janz"}

	args = append(args, params.asArgs()...)

	DEBUG.Printf("Starting tc: %+v", args)
	res, err := commands.ExecReturnOutput("tc", args...)
	if err != nil {
		INFO.Printf("'tc %+v' -> %s", args, res)
		return err
	}
	tc.control_file, err = os.OpenFile(CTRL_FILE, os.O_WRONLY, os.ModeAppend)
	return err
}
func (tc *TrafficControl) CurrentRate() float64 {
	return tc.current_data_rate
}
func (tc *TrafficControl) Reset() error {
	_, err := commands.ExecReturnOutput("tc", "qdisc", "delete", "dev", tc.dev, "root")
	return err
}

func (tc *TrafficControl) Close() error {
	DEBUG.Println("Closing tc")
	if tc.nft.L4sPremarking {
		ResetECTMarking(assets.NFT_TABLE_PREMARK)
	}
	if tc.nft.SignalStart {
		ResetECTMarking(assets.NFT_TABLE_SIGNAL)
	}

	if err := tc.Reset(); err != nil {
		INFO.Printf("Could not reset TC in Close(): %v", err)
	}
	if err := tc.control_file.Close(); err != nil {
		INFO.Printf("Could not close control_file TC in Close(): %v", err)
	}
	DEBUG.Println("Closed tc")
	return nil
}

func (tc *TrafficControl) ChangeTo(rate float64) error {
	changeRateArray := make([]byte, 8)
	currentDataRateBit := uint64(rate) * 1000
	tc.current_data_rate = rate
	binary.LittleEndian.PutUint64(changeRateArray, currentDataRateBit)
	_, err := tc.control_file.Write(changeRateArray)
	return err
}

func (tc *TrafficControl) LaunchChangeLoop(waitTime time.Duration, wg *sync.WaitGroup, drp *datatypes.DB_data_rate_pattern) error {
	defer wg.Done()
	ticker := time.NewTicker(waitTime)
	INFO.Println("start playing DataRatePattern")
	if tc.nft.SignalStart {
		ResetECTMarking(assets.NFT_TABLE_SIGNAL)
		CreateNftRuleECT(tc.dev, assets.NFT_TABLE_SIGNAL, assets.NFT_CHAIN_FORWARD, assets.NFT_CHAIN_OUTPUT, "ect0", "1")
		go func() {
			<-time.NewTimer(200 * time.Millisecond).C
			ResetECTMarking(assets.NFT_TABLE_SIGNAL)
		}()
	}
	// open controll file for rate changes

	for range ticker.C {
		// perhaps signal start of drp to clients by marking with ECT0
		value, err := drp.Next()
		if err != nil {
			if _, ok := err.(*errortypes.IterableStopError); ok {
				DEBUG.Println("Pattern finished")
				break
			} else {
				FATAL.Println(err)
			}
		}
		//change data rate in controll file
		if err := tc.ChangeTo(value); err != nil {
			return err
		}
	}
	return nil
}
