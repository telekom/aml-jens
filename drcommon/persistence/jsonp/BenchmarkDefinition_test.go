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
	"jens/drcommon/persistence/jsonp"
	"jens/drcommon/util/utiltest"
	"testing"
)

func TestBenchmarkDefinitionJsonMarshall(t *testing.T) {
	data := jsonp.BenchmarkDefinition{}
	txtbin, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	txt := string(txtbin)
	utiltest.InJsonOutput(t, txt, "Name")
	utiltest.InJsonOutput(t, txt, "Max_application_bitrate")
	utiltest.InJsonOutput(t, txt, "MaxBitrateEstimationTimeS")
	utiltest.InJsonOutput(t, txt, "Patterns")
	utiltest.InJsonOutput(t, txt, "DrplaySetting")
}
