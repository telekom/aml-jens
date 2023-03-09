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
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"

	"github.com/spf13/viper"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

// Modifies c.player and c.shower
func (c *config) setDefaults() {
	//Maybe the optimizer does this if I use string.split but
	//this is fine.
	c.benchmark = &DrBenchmarkConfig{}
	c.player.A_Session.Time = uint64(time.Now().UnixMilli())
}

// Creates ConfigObjects in Memory and reads all
// config settings from the files specified.
func (c *config) readFromFile() error {
	viper.SetConfigName(assets.CFG_FILE_NAME)
	viper.AddConfigPath(paths.RELEASE_CFG_PATH()) // for production
	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		return err
	}
	INFO.Printf("Config used: %s", viper.ConfigFileUsed())
	drp := datatypes.NewDB_data_rate_pattern()
	drp.Intial_minRateKbits = viper.GetFloat64("drp.minRateKbits")
	drp.WarmupTimeMs = viper.GetInt32("drp.WarmupBeforeDrpMs")
	drp.Freq = -1
	drp.Initial_scale = -1
	asd := datatypes.DB_session{
		//TC
		Markfree:            viper.GetInt32("tccommands.markfree"),
		Markfull:            viper.GetInt32("tccommands.markfull"),
		Queuesizepackets:    viper.GetInt32("tccommands.queuesizepackets"),
		ExtralatencyMs:      viper.GetInt32("tccommands.extralatency"),
		L4sEnablePreMarking: viper.GetBool("tccommands.l4sEnabledPreMarking"),
		SignalDrpStart:      viper.GetBool("tccommands.signalDrpStart"),
		//DRP
		ChildDRP: drp,
		ParentBenchmark: &datatypes.DB_benchmark{
			PrintToStdOut: true,
		},
	}
	c.player = &DrPlayConfig{
		A_Session: &asd,
		Psql: datatypes.Login{
			Dbname:   viper.GetString("postgres.dbname"),
			Host:     viper.GetString("postgres.host"),
			Password: viper.GetString("postgres.password"),
			Port:     viper.GetInt32("postgres.port"),
			User:     viper.GetString("postgres.user"),
		},
		PrintToStdOut: true,
	}
	var scaleMode ConfigFlowPlotScrollMode = Scrolling
	if viper.GetBool("drshow.scalePlots") {
		scaleMode = Scaling
	}
	c.shower = &showConfig{
		PlotScrollMode:          scaleMode,
		WaitTimeForSamplesInSec: (viper.GetInt32("drp.WarmupBeforeDrpMs") / 1000) + 2,
		ExportPathPrefix:        viper.GetString("drshow.exportPath"),
		FlowEligibility: showConfigFlowEligibilityT{
			Level0Until: cfgIntDefault("drshow.FilterLevel0Until", 7),
			Level1: showConfigFlowEligibilityEntryT{
				AdditionalCount:   cfgIntDefault("drshow.FilterLevel1AddUntil", 6),
				MinimumFlowLength: cfgIntDefault("drshow.FilterLevel1MinSamples", 15),
				TimePastMaxInSec:  cfgIntDefault("drshow.FilterLevel1SecsPassed", 60),
			},
			Level2: showConfigFlowEligibilityEntryT{
				AdditionalCount:   cfgIntDefault("drshow.FilterLevel2AddUntil", 6),
				MinimumFlowLength: cfgIntDefault("drshow.FilterLevel2MinSamples", 30),
				TimePastMaxInSec:  cfgIntDefault("drshow.FilterLevel2SecsPassed", 40),
			},
			Level3: showConfigFlowEligibilityEntryT{
				AdditionalCount:   0,
				MinimumFlowLength: cfgIntDefault("drshow.FilterLevel3MinSamples", 40),
				TimePastMaxInSec:  cfgIntDefault("drshow.FilterLevel3SecsPassed", 20),
			},
		},
	}
	return nil
}

// Set internal config to some default values
// this will only be called if there is no config
// present in any specified directory.
func (c *config) setToDefaults() {
	drp := datatypes.NewDB_data_rate_pattern()
	drp.Intial_minRateKbits = 500
	drp.WarmupTimeMs = 2000
	drp.Freq = -1
	drp.Initial_scale = -1
	asd := datatypes.DB_session{
		//TC
		Markfree:            4,
		Markfull:            14,
		Queuesizepackets:    10000,
		ExtralatencyMs:      0,
		L4sEnablePreMarking: false,
		SignalDrpStart:      false,
		//DRP
		ChildDRP: drp,
		ParentBenchmark: &datatypes.DB_benchmark{
			PrintToStdOut: true,
		},
	}
	c.player = &DrPlayConfig{
		A_Session: &asd,
		Psql: datatypes.Login{
			Dbname:   "default",
			Host:     "default",
			Password: "default",
			Port:     0,
			User:     "default",
		},
		PrintToStdOut: true,
	}
	var scaleMode ConfigFlowPlotScrollMode = Scrolling
	scaleMode = Scaling
	c.shower = &showConfig{
		PlotScrollMode:          scaleMode,
		WaitTimeForSamplesInSec: (2000 / 1000) + 2,
		ExportPathPrefix:        "/tmp/jens/",
		FlowEligibility: showConfigFlowEligibilityT{
			Level0Until: 7,
			Level1: showConfigFlowEligibilityEntryT{
				AdditionalCount:   6,
				MinimumFlowLength: 15,
				TimePastMaxInSec:  60,
			},
			Level2: showConfigFlowEligibilityEntryT{
				AdditionalCount:   6,
				MinimumFlowLength: 30,
				TimePastMaxInSec:  40,
			},
			Level3: showConfigFlowEligibilityEntryT{
				AdditionalCount:   0,
				MinimumFlowLength: 40,
				TimePastMaxInSec:  20,
			},
		},
	}
}

func cfgIntDefault(key string, def int) int {
	res := viper.GetInt(key)
	if res == 0 {
		return def
	}
	return res
}
