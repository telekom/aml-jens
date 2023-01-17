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
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

const DRP3VALL = "drp_3valleys.csv"

var GOOD_PATH = filepath.Join(paths.TESTDATA_DRP(), DRP3VALL)
var SAW_PATH = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")

func TestDrpNaming(t *testing.T) {
	drp := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 500,
		Loop:         false,
	}
	err := drp.ParseDrpFile(GOOD_PATH)
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	expected := DRP3VALL
	if is := drp.Name; is != expected {
		t.Fatalf("Name of DRP set incorrectly: is '%s', should be %s", is, expected)
	}
	if drp.IsLooping() != false {
		t.Fatalf("Pattern is incorrectly set to loop=true")
	}
}

func TestDrpMinrateAndLen(t *testing.T) {
	drp := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 20400,
		Loop:         false,
	}
	err := drp.ParseDrpFile(GOOD_PATH)
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	counter := 0
	for {
		rate, err := drp.Next()
		if err != nil {
			if _, ok := err.(*errortypes.IterableStopError); !ok {
				t.Fatal(err)
			}
			if expected := 3000; counter != expected {
				t.Fatalf("Length of loaded pattern is incorrect (%d != %d)", counter, expected)
			}
			return
		}
		if rate < 20400 {
			t.Fatalf("Encountered Rate less than set MinimumBitRate: %f", rate)
		}
		counter++
	}
}

func TestDrpHash(t *testing.T) {
	drp := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 20400,
		Loop:         false,
	}
	err := drp.ParseDrpFile(GOOD_PATH)
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	if drp.GetHashStr() != "020b6fc00d7a6f91c050a0833f11c18d" {
		t.Fatalf("Hash is incorrect. Is: '%s', should be %s", drp.GetHashStr(), "310f51cd6764b358bc1af15c53d138e5")
	}
}

func TestBroken(t *testing.T) {
	files, err := ioutil.ReadDir(paths.TESTDATA_DRP())
	if err != nil {
		t.Fatalf("While reading testdata: %s", err)
	}
	for _, v := range files {
		if !strings.HasPrefix(v.Name(), "broken_") {
			continue
		}

		path := filepath.Join(paths.TESTDATA_DRP(), v.Name())
		drp := datatypes.DB_data_rate_pattern{
			Scale:        1,
			MinRateKbits: 20400,
			Loop:         false,
		}
		err := drp.ParseDrpFile(path)
		if err == nil {
			t.Fatalf("Loaded Invalid Pattern '%s' -> %+v", path, drp)
		}

	}

}

func TestDrpValueNoLoop(t *testing.T) {
	var expected = []float64{10000, 20000, 30000, 40000, 50000, 60000, 70000, 80000, 90000, 100000}
	drp := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 600,
		Loop:         false,
	}
	err := drp.ParseDrpFile(SAW_PATH)
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	counter := 0
	for {
		rate, err := drp.Next()
		if err != nil {
			if _, ok := err.(*errortypes.IterableStopError); !ok {
				t.Fatal(err)
			}
			if expected := 10; counter != expected {
				t.Fatalf("Length of loaded pattern is incorrect (%d != %d)", counter, expected)

			}
			return
		}
		if rate != expected[counter] {
			t.Fatalf("Rate(%f) @%d != expected %f", rate, counter, expected[counter])
		}
		counter++
	}
}
func TestDrpValueLoop(t *testing.T) {
	expected := []float64{
		10000, 20000, 30000, 40000, 50000, 60000, 70000, 80000, 90000, 100000,
		100000, 90000, 80000, 70000, 60000, 50000, 40000, 30000, 20000, 10000,
		10000, 20000, 30000, 40000, 50000, 60000, 70000, 80000, 90000, 100000,
		100000, 90000, 80000, 70000, 60000, 50000, 40000, 30000, 20000, 10000,
		10000, 20000, 30000, 40000, 50000, 60000, 70000, 80000, 90000, 100000,
	}
	drp := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 600,
		Loop:         true,
	}
	err := drp.ParseDrpFile(SAW_PATH)
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	counter := 0
	for {
		rate, err := drp.Next()
		if err != nil {
			t.Fatalf("Got %v while looping", err)
		}
		//t.Logf("%02d: %f", counter, rate)
		if rate != expected[counter] {
			t.Fatalf("Got %f, expected %f", rate, expected[counter])
		}
		counter++
		if counter == len(expected) {
			return
		}
	}
}
