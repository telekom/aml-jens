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

package config

import (
	"jens/drcommon/util/utiltest"
	"testing"

	"github.com/spf13/viper"
)

var TEST_CONFIG = `
[tccommands]
  markfree=9 #in Millisecond
  markfull=90 #in Millisecond
  # Size of custom Queue in packets
  queuesizepackets=10000
  # AddonLatency in MS to add to all packets
  extralatency=10
  # Mark non-ect(1) Traffic as enabled
  l4sEnabledPreMarking=true
  # Mark the first packets with special ect
  signalDrpStart=true

[postgres]
  dbname = "dbname"
  host = "host"
  password = "password"
  port = 12345
  user = "user"

[drp]
  # MinimumDataRate for patterns
  minRateKbits = 1
  #Phase in MS before Networkshaping /DRP takes effect
  WarmupBeforeDrpMs = 1000

[drshow]
  scalePlots=true #instead of scrolling
  exportPath="/etc/jens-cli"
  #Used for filtering of flows in pipemode
  FilterLevel0Until=1     
  #Adds to Level0Until: Min count of flows for filter to kick in action
  FilterLevel1AddUntil=1   
  #Filter minium samples to display flow
  FilterLevel1MinSamples=1
  #Or maximum time passed since last sample was received
  FilterLevel1SecsPassed=1
  FilterLevel2AddUntil=2
  FilterLevel2MinSamples=2
  FilterLevel2SecsPassed=2
  FilterLevel3AddUntil=3
  FilterLevel3MinSamples=3
  FilterLevel3SecsPassed=3

`

func TestConfig(t *testing.T) {
	viper.AddConfigPath("/tmp/") // for Tests
	defer utiltest.PreTestWriteFile(t, "/tmp/config.toml", []byte(TEST_CONFIG))()
	c, err := createNewConfig()
	if err != nil {
		t.Fatal(err)
	}
	player := []bool{
		c.player.A_Session.Markfree == 9,
		c.player.A_Session.Markfull == 90,
		c.player.A_Session.L4sEnablePreMarking == true,
		c.player.A_Session.SignalDrpStart == true,
		c.player.A_Session.ExtralatencyMs == 10,
		c.player.A_Session.ChildDRP.WarmupTimeMs == 1000,
		c.player.A_Session.ChildDRP.MinRateKbits == 1,
	}
	other := []bool{
		c.player.Psql.Dbname == "dbname",
		c.player.Psql.Host == "host",
		c.player.Psql.Password == "password",
		c.player.Psql.Port == 12345,
		c.player.Psql.User == "user",
		c.player.PrintToStdOut == true,
	}
	s := c.shower
	shower := []bool{
		s.PlotScrollMode == 1,
		s.FlowEligibility.Level0Until == 1,
		s.FlowEligibility.Level1.AdditionalCount == 1,
		s.FlowEligibility.Level1.MinimumFlowLength == 1,
		s.FlowEligibility.Level1.TimePastMaxInSec == 1,

		s.FlowEligibility.Level2.AdditionalCount == 2,
		s.FlowEligibility.Level2.MinimumFlowLength == 2,
		s.FlowEligibility.Level2.TimePastMaxInSec == 2,
		s.FlowEligibility.Level3.MinimumFlowLength == 3,
		s.FlowEligibility.Level3.TimePastMaxInSec == 3,
	}

	for i, v := range player {
		if !v {
			t.Log(i)
			t.Fatalf("Loading DrPlay-Session config did not work:\n\t%+v\n\t%+v", c.player, player)
			return
		}
	}
	for i, v := range other {
		if !v {
			t.Log(i)
			t.Fatalf("Loading DrPlay-other config did not work:\n\t%+v\n\t%+v", c.shower, shower)
			return
		}
	}
	for i, v := range shower {
		if !v {
			t.Log(i)
			t.Fatalf("Loading DrShow config did not work:\n\t%+v\n\t%+v", c.shower, shower)
			return
		}

	}

}
