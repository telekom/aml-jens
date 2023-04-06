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

package jsonp_test

import (
	"encoding/json"
	"testing"

	"github.com/telekom/aml-jens/internal/persistence/jsonp"
	"github.com/telekom/aml-jens/internal/util/utiltest"
)

func TestNewBenchmarkDrPlaySetting(t *testing.T) {
	data := jsonp.NewDrplaySetting(1, 0.1, 100, 12000)
	if *data.DRP.Frequency != 1 {
		t.Fatalf("Frequency was not set in constructor")
	}
	if *data.DRP.Scale != 0.1 {
		t.Fatalf("Scale was not set in constructor")
	}
	if *data.DRP.MinRateKbits != 100 {
		t.Fatalf("MinRateKbits was not set in constructor")
	}
	if *data.DRP.WarmupBeforeDrpMs != 12000 {
		t.Fatalf("WarmupBeforeDrpMs was not set in constructor")
	}
}
func TestBenchmarkDrPlaySettingJsonMarshall(t *testing.T) {
	data := jsonp.NewDrplaySetting(1, 0.1, 100, 12000)
	txtbin, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	txt := string(txtbin)
	utiltest.InJsonOutput(t, txt, "MinRateKbits")
	utiltest.InJsonOutput(t, txt, "Scale")
	utiltest.InJsonOutput(t, txt, "Frequency")
	utiltest.InJsonOutput(t, txt, "WarmupBeforeDrpMs")
}
