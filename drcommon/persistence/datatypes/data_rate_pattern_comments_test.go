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

package datatypes_test

import (
	"jens/drcommon/assets"
	"jens/drcommon/persistence/datatypes"
	"testing"
)

func TestGenericComment(t *testing.T) {
	data := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 20400,
		Loop:         false,
	}
	err := data.ParseDrpFile(assets.TESTDATA_DPR_DIR + "drp_3valleys_generic_comment.csv")
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	if data.GetHashStr() != "020b6fc00d7a6f91c050a0833f11c18d" {
		t.Fatalf("Hash is incorrect. Is: '%s', should be %s", data.GetHashStr(), "310f51cd6764b358bc1af15c53d138e5")
	}
	if data.Description.String == "" {
		t.Fatal("Description field was not set correctly!")
	}
}

func TestKeyValueComment(t *testing.T) {
	data := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 20400,
		Loop:         false,
	}
	err := data.ParseDrpFile(assets.TESTDATA_DPR_DIR + "drp_3valleys_kv_comment.csv")
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	if data.GetHashStr() != "020b6fc00d7a6f91c050a0833f11c18d" {
		t.Fatalf("Hash is incorrect. Is: '%s', should be %s", data.GetHashStr(), "310f51cd6764b358bc1af15c53d138e5")
	}
	if !data.Description.Valid || data.Description.String == "" {
		t.Fatal("Description field was not set correctly!")
	}
	if data.Th_mq_latency != "{3,6}" ||
		data.Th_p95_latency != "{11,21}" ||
		data.Th_p99_latency != "{11,21}" ||
		data.Th_p999_latency != "{11,21}" ||
		data.Th_link_usage != "{90,80}" {
		t.Logf("Th_mq_latency: %+v", data.Th_mq_latency)
		t.Logf("Th_p95_latency: %+v", data.Th_p95_latency)
		t.Logf("Th_p99_latency: %+v", data.Th_p99_latency)
		t.Logf("Th_p999_latency: %+v", data.Th_p999_latency)
		t.Logf("Th_link_usage: %+v", data.Th_link_usage)
		t.Fatal("DataRatePatternEvaluation did not get set correctly")
	}
}
