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
	"strconv"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/commands"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"

	"os"
	"time"

	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/logging"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

const JANZ_CTRL_FILE = "/sys/kernel/debug/sch_janz/0001:v1"
const MULTIJENS_CTRL_FILE = "/sys/kernel/debug/sch_multijens/0001:v1"

type TrafficControlStartParams struct {
	Datarate     uint32
	QueueSize    int
	AddonLatency int
	Markfree     int
	Markfull     int
	Qosmode      uint8
	Uenum        uint8
}
type NftStartParams struct {
	L4sPremarking bool
	SignalStart   bool
	Netflows      []string
	Uenum         uint8
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
	if p.Qosmode < 0 || p.Qosmode > 2 {
		return errortypes.NewUserInputError("valid values for qosmode are 0,1,2")
	}
	if p.Uenum < 0 || p.Uenum > 16 {
		return errortypes.NewUserInputError("valid values for uenum are in [1..16]")
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
	if p.Qosmode > 0 {
		args = append(args, "qosmode", fmt.Sprintf("%d", p.Qosmode))
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

// Init sets NFT and TC to workable state, connects to custom qdisk.
// After calling Init Close has to be called.
func (tc *TrafficControl) Init(params TrafficControlStartParams, nft NftStartParams) error {
	ResetNFT(assets.NFT_TABLE_PREMARK)
	if nft.L4sPremarking {
		err := CreateRuleECT(tc.dev, assets.NFT_TABLE_PREMARK, assets.NFT_CHAIN_FORWARD, assets.NFT_CHAIN_OUTPUT, "ect1", "0")
		if err != nil {
			return err
		}
	}

	if err := params.validate(); err != nil {
		return err
	}
	if err := tc.Reset(); true {
		DEBUG.Printf("TcReset: %v", err)
	}
	args := []string{"qdisc", "add", "dev", tc.dev, "root", "handle", "1:", "janz"}

	args = append(args, params.asArgs()...)
	time.Sleep(1 * time.Second)
	DEBUG.Printf("Starting tc: %+v", args)
	res := commands.ExecCommand("tc", args...)
	if res.Error() != nil {
		return res.Error()
	}
	var err error
	tc.control_file, err = os.OpenFile(JANZ_CTRL_FILE, os.O_WRONLY, os.ModeAppend)
	return err
}

// Init sets NFT and TC to workable state, connects to custom qdisk.
// After calling Init Close has to be called.
func (tc *TrafficControl) InitMultijens(params TrafficControlStartParams, nft NftStartParams) error {
	ResetNFT(assets.NFT_TABLE_PREMARK)
	tc.nft = nft

	if nft.L4sPremarking {
		err := CreateRuleECT(tc.dev, assets.NFT_TABLE_PREMARK, assets.NFT_CHAIN_FORWARD, assets.NFT_CHAIN_OUTPUT, "ect1", "0")
		if err != nil {
			return err
		}
	}
	// create nft mark rules for queue assignment
	ResetNFT(assets.NFT_TABLE_UEMARK)
	err := CreateRulesMarkUe(tc.nft.Netflows)
	if err != nil {
		return err
	}
	if err := params.validate(); err != nil {
		return err
	}
	if err := tc.Reset(); true {
		DEBUG.Printf("TcReset: %v", err)
	}
	var args = []string{"qdisc", "add", "dev", tc.dev, "root", "handle", "1:", "multijens", "uenum", strconv.FormatUint(uint64(params.Uenum), 10)}

	args = append(args, params.asArgs()...)
	time.Sleep(1 * time.Second)
	DEBUG.Printf("Starting tc: %+v", args)
	res := commands.ExecCommand("tc", args...)
	if res.Error() != nil {
		return res.Error()
	}
	tc.control_file, err = os.OpenFile(MULTIJENS_CTRL_FILE, os.O_WRONLY, os.ModeAppend)
	return err
}

// Rests Qdisc to default
func (tc *TrafficControl) Reset() error {
	return commands.ExecCommand("tc", "qdisc", "delete", "dev", tc.dev, "root").Error()
}

// Closes all open contexts; Resets NFT_TABLE, tc markings etc.
//
// This function needs to be called after tc is Done.
func (tc *TrafficControl) Close() error {
	DEBUG.Println("Closing tc")
	if tc.nft.L4sPremarking {
		ResetNFT(assets.NFT_TABLE_PREMARK)
	}
	if tc.nft.SignalStart {
		ResetNFT(assets.NFT_TABLE_SIGNAL)
	}

	_ = tc.Reset()
	if err := tc.control_file.Close(); err == nil {
		//This is to be expected: File gets closed beforehand
		WARN.Printf("control_file TC had to be closed")
	} else {
		DEBUG.Printf("Could not close control_file TC, %s", err)
	}
	return nil
}

// Changes the current bandwidth limit to rate
func (tc *TrafficControl) ChangeTo(rate float64) error {
	changeRateArray := make([]byte, 8)
	currentDataRateBit := uint64(rate) * 1000
	tc.current_data_rate = rate
	binary.LittleEndian.PutUint64(changeRateArray, currentDataRateBit)
	_, err := tc.control_file.Write(changeRateArray)
	return err
}

// Changes the current bandwidth limit to rate
func (tc *TrafficControl) ChangeMultiTo(rate float64) error {
	changeRateArray := make([]byte, 8*tc.nft.Uenum)
	currentDataRateBit := uint64(rate) * 1000
	tc.current_data_rate = rate
	for i := 0; i < int(tc.nft.Uenum); i++ {
		offset := i * 8
		binary.LittleEndian.PutUint64(changeRateArray[offset:], currentDataRateBit)
	}
	_, err := tc.control_file.Write(changeRateArray)
	return err
}

// Starts a goroutine that will change the current bandwidth restriciton.
// A change will occur after the waitTime is exceeded.
//
// # Uses util.RoutineReport
//
// Blockig - also spawns 1 short lived routine
func (tc *TrafficControl) LaunchChangeLoop(waitTime time.Duration, drp *datatypes.DB_data_rate_pattern, r util.RoutineReport) {
	ticker := time.NewTicker(waitTime)
	INFO.Printf("start playing DataRatePattern @%s", waitTime.String())
	if tc.nft.SignalStart {
		go func() {
			ResetNFT(assets.NFT_TABLE_SIGNAL)
			err := CreateRuleECT(tc.dev, assets.NFT_TABLE_SIGNAL, assets.NFT_CHAIN_FORWARD, assets.NFT_CHAIN_OUTPUT, "ect0", "1")
			if err != nil {
				r.ReportFatal(err)
				return
			}
			<-time.NewTimer(200 * time.Millisecond).C
			ResetNFT(assets.NFT_TABLE_SIGNAL)
		}()
	}
	for {
		select {
		case <-r.On_extern_exit_c:
			DEBUG.Println("Closing TC-loop")
			r.Wg.Done()
			return
		case <-ticker.C:
			value, err := drp.Next()
			if err != nil {
				if _, ok := err.(*errortypes.IterableStopError); ok {
					r.Application_has_finished <- "DataRatePattern has finished"
					r.Wg.Done()
					return
				} else {
					r.ReportWarn(fmt.Errorf("LaunchChangeLoop could retrieve next Value: %w", err))
					r.Wg.Done()
					return
				}
			}
			//change data rate in control file
			if err := tc.ChangeTo(value); err != nil {
				r.ReportFatal(fmt.Errorf("LaunchChangeLoop could not change Value: %w", err))
				r.Wg.Done()
				return
			}
		}
	}
}

// Starts a goroutine that will change the current capacity restricitons.
// A change will occur after the waitTime is exceeded.
//
// # Uses util.RoutineReport
//
// Blockig - also spawns 1 short lived routine
func (tc *TrafficControl) LaunchMultiChangeLoop(waitTime time.Duration, drp *datatypes.DB_data_rate_pattern, r util.RoutineReport) {
	ticker := time.NewTicker(waitTime)
	INFO.Printf("start playing DataRatePattern @%s", waitTime.String())
	for {
		select {
		case <-r.On_extern_exit_c:
			DEBUG.Println("Closing TC-loop")
			r.Wg.Done()
			return
		case <-ticker.C:
			value, err := drp.Next()
			if err != nil {
				if _, ok := err.(*errortypes.IterableStopError); ok {
					r.Application_has_finished <- "DataRatePattern has finished"
					r.Wg.Done()
					return
				} else {
					r.ReportWarn(fmt.Errorf("LaunchChangeLoop could retrieve next Value: %w", err))
					r.Wg.Done()
					return
				}
			}
			//change data rate in control file
			if err := tc.ChangeMultiTo(value); err != nil {
				r.ReportFatal(fmt.Errorf("LaunchChangeLoop could not change value: %w", err))
				r.Wg.Done()
				return
			}
		}
	}
}
