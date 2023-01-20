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
	"path/filepath"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/pkg/drp"
)

func TestGenericComment(t *testing.T) {
	data := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 20400,
	}
	p := paths.TESTDATA_DRP()
	path := filepath.Join(p, "drp_3valleys_generic_comment.csv")

	err := data.ParseDRP(drp.NewDataRatePatternFileProvider(path))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	if data.GetHashStr() != "020b6fc00d7a6f91c050a0833f11c18d" {
		t.Fatalf("Hash is incorrect. Is: '%s', should be %s", data.GetHashStr(), "310f51cd6764b358bc1af15c53d138e5")
	}
	if data.GetDescription().String == "" {
		t.Fatal("Description field was not set correctly!")
	}
}

func TestKeyValueComment(t *testing.T) {
	data := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 20400,
	}
	err := data.ParseDRP(drp.NewDataRatePatternFileProvider(filepath.Join(paths.TESTDATA_DRP(), "drp_3valleys_kv_comment.csv")))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	if data.GetHashStr() != "020b6fc00d7a6f91c050a0833f11c18d" {
		//if scaling is applied: 310f51cd6764b358bc1af15c53d138e5
		t.Fatalf("Hash is incorrect. Is: '%s', should be %s", data.GetHashStr(), "020b6fc00d7a6f91c050a0833f11c18d")
	}
	if !data.GetDescription().Valid || data.GetDescription().String == "" {
		t.Fatal("Description field was not set correctly!")
	}
	if data.GetTh_mq_latency() != "{3,6}" ||
		data.GetTh_p95_latency() != "{11,21}" ||
		data.GetTh_p99_latency() != "{11,21}" ||
		data.GetTh_p999_latency() != "{11,21}" ||
		data.GetTh_link_usage() != "{90,80}" {
		t.Logf("Th_mq_latency: %+v", data.GetTh_mq_latency())
		t.Logf("Th_p95_latency: %+v", data.GetTh_p95_latency())
		t.Logf("Th_p99_latency: %+v", data.GetTh_p99_latency())
		t.Logf("Th_p999_latency: %+v", data.GetTh_p999_latency())
		t.Logf("Th_link_usage: %+v", data.GetTh_link_usage())
		t.Fatal("DataRatePatternEvaluation did not get set correctly")
	}
}
