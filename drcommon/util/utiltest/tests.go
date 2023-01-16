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

package utiltest

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var TEST_CONFIG_PATH = "./testdata/config/"

const PATTERN_ORIGIN_PATH = "../testdata/drp_3valleys.csv"
const DRP_PATH = "./drp_3valleys.csv"

func failPrepD1(format string, args ...interface{}) string {
	_, file, line, _ := runtime.Caller(3)
	return fmt.Sprintf("\t%s:%d: %s\n", filepath.Base(file), line, fmt.Sprintf(format, args...))

}

func NotInJsonOutput(t *testing.T, json string, keyword string) {
	if strings.Contains(json, keyword) {
		t.Fatalf(failPrepD1("Keyword '%s' in Json output", keyword))
	}
}

func InJsonOutput(t *testing.T, json string, keyword string) {
	if !strings.Contains(json, keyword) {
		t.Fatalf(failPrepD1("Keyword '%s' not in Json output", keyword))
	}
}

func PreTestCreateConfig(t *testing.T) func() {
	const config_data = `
	[tccommands]
	  markfree=9 #in Millisecond
	  markfull=90 #in Millisecond
      queuesizepackets=10000
      extralatency=10
      l4sEnabledPreMarking=true
	  signalDrpStart=true
	
	[postgres]
	  dbname = "dbname"
	  host = "host"
	  password = "password"
	  port = 12345
	  user = "user"
	
	[drp]
	  minRateKbits = 1
	  WarmupBeforeDrpMs = 9999
	
	[drshow]
	  scalePlots=true #instead of scrolling
	  exportPath="/tmp"
	  FilterLevel0Until=1       #Used for filtering of flows in pipemode
	  FilterLevel1AddUntil=1    #Adds to Level0Until: Min count of flows for filter to kick in action
	  FilterLevel1MinSamples=1 #Filter minium samples to display flow
	  FilterLevel1SecsPassed=1 #Or maximum time passed since last sample was received
	  FilterLevel2AddUntil=2
	  FilterLevel2MinSamples=2
	  FilterLevel2SecsPassed=2
	  FilterLevel3MinSamples=3
	  FilterLevel3SecsPassed=3
		
	`
	return PreTestWriteFile(t, TEST_CONFIG_PATH+"config.toml", []byte(config_data))
}

func PreTestCopyFile(t *testing.T, origin string, dest string) func() {
	data, err := os.ReadFile(origin)
	if err != nil {
		t.Fatal(err)
	}
	return PreTestWriteFile(t, dest, data)
}

func PreTestWriteFile(t *testing.T, dest string, data []byte) func() {
	if err := os.WriteFile(dest, data, 0777); err != nil {
		t.Fatal(err)
	}
	return func() { os.Remove(dest) }
}
