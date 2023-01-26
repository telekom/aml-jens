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
package drp

import (
	"path/filepath"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
)

func TestReadCSV_OK_saw(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")
	data, err := readCSV(path)
	t.Log(data)
	if err != nil {
		t.Fatal(err)
	}
}
func TestConvertDRPdata_OK_saw(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")
	strdata, err := readCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := convertDRPdata(strdata, struct {
		MinRateKbits float64
		Scale        float64
		Origin       string
	}{
		Scale:        1,
		MinRateKbits: 0,
		Origin:       "Test: " + path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if data.Length != 10 {
		t.Fatal("Incorrect amount of entries loaded")
	}
	if data.Max != 100000 {
		t.Fatalf("Incorrect stat: Max is %f should be 100000", data.Max)
	}
	if data.Min != 10000 {
		t.Fatalf("Incorrect stat: Min is %f should be 10000", data.Min)
	}
	if data.Avg != 55000 {
		t.Fatalf("Incorrect stat: Avg is %f should be 55000", data.Avg)
	}
	expected := []float64{
		10000,
		20000,
		30000,
		40000,
		50000,
		60000,
		70000,
		80000,
		90000,
		100000,
	}
	if len(*data.data) != len(expected) &&
		(*data.data)[0] != expected[0] &&
		(*data.data)[1] != expected[1] &&
		(*data.data)[2] != expected[2] &&
		(*data.data)[3] != expected[3] &&
		(*data.data)[4] != expected[4] &&
		(*data.data)[5] != expected[5] &&
		(*data.data)[6] != expected[6] &&
		(*data.data)[7] != expected[7] &&
		(*data.data)[8] != expected[8] &&
		(*data.data)[9] != expected[9] {
		t.Log(data.data)
		t.Fatal("drp-data was not loaded correctly")
	}

}

func Validate3Valleys_Stats(t *testing.T, data DataRatePattern) {
	t.Helper()
	if data.Length != 3000 {
		t.Fatal("Incorrect amount of entries loaded")
	}
	if data.Max != 20998 {
		t.Fatalf("Incorrect stat: Max is %f should be 20998", data.Max)
	}
	if data.Min != 5053 {
		t.Fatalf("Incorrect stat: Min is %f should be 5053", data.Min)
	}
	if data.Avg != 16818.511 {
		t.Fatalf("Incorrect stat: Avg is %f should be 16818.511", data.Avg)
	}
}

func TestDataRatePatternFileProvider_OK_3valleys(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "drp_3valleys.csv")
	data, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	Validate3Valleys_Stats(t, data)
}
func TestDataRatePatternFileProvider_OK_3valleys_generic_comment(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "drp_3valleys_generic_comment.csv")
	data, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	Validate3Valleys_Stats(t, data)
	expected := `This is a generic comment
I expect this to also count as a comment
!"§$%&/()=?'#*+~<>|
¹²³¼½¬{[]}
ÄÖÜ€µ
`
	if data.Description != expected {
		t.Log("Got:")
		t.Log(data.Description)
		t.Log("Expected:")
		t.Log(expected)
		t.Fatal("Generic comment was incorrectly read")
	}
}

func TestDataRatePatternFileProvider_OK_3valleys_kv_comment(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "drp_3valleys_kv_comment.csv")
	data, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	Validate3Valleys_Stats(t, data)
	expected := `This is a generic comment
I expect this to also count as a comment
`
	if data.Description != expected {
		t.Log("Got:")
		t.Log(data.Description)
		t.Log("Expected:")
		t.Log(expected)
		t.Fatal("Generic comment was incorrectly read")
	}

	v, found := data.mapping["th_mq_latency"]
	if !found {
		t.Fatalf("Key 'th_mq_latency' not found")
	}
	if v != "{3,6}" {
		t.Fatalf("th_mq_latency!=3,6; Got: %s", v)
	}
	v, found = data.mapping["th_p95_latency"]
	if !found {
		t.Fatalf("Key 'th_p95_latency' not found")
	}
	if v != "{11,21}" {
		t.Fatalf("th_p95_latency!=11,21; Got: %s", v)
	}
	v, found = data.mapping["th_p99_latency"]
	if !found {
		t.Fatalf("Key 'th_p99_latency' not found")
	}
	if v != "{11,21}" {
		t.Fatalf("th_p99_latency!=11,21; Got: %s", v)
	}
	v, found = data.mapping["th_p999_latency"]
	if !found {
		t.Fatalf("Key 'th_p999_latency' not found")
	}
	if v != "{11,21}" {
		t.Fatalf("th_p999_latency!=11,21; Got: %s", v)
	}
	v, found = data.mapping["th_link_usage"]
	if !found {
		t.Fatalf("Key 'th_link_usage' not found")
	}
	if v != "{90,80}" {
		t.Fatalf("th_link_usage!=90,80; Got: %s", v)
	}
}
