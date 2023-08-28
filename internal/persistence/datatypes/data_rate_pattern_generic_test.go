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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/pkg/drp"
)

const DRP3VALL = "drp_3valleys.csv"

var GOOD_PATH = filepath.Join(paths.TESTDATA_DRP(), DRP3VALL)
var SAW_PATH = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")

func TestDrpNaming(t *testing.T) {
	db_drp := datatypes.DB_data_rate_pattern{
		Initial_scale:       1,
		Intial_minRateKbits: 500,
	}
	err := db_drp.ParseDRP(drp.NewDataRatePatternFileProvider(GOOD_PATH))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	expected := DRP3VALL
	if is := db_drp.GetName(); is != expected {
		t.Fatalf("Name of DRP set incorrectly: is '%s', should be %s", is, expected)
	}
	if db_drp.IsLooping() != false {
		t.Fatalf("Pattern is incorrectly set to loop=true")
	}
}

func TestDrpMinrateAndLen(t *testing.T) {
	db_drp := datatypes.DB_data_rate_pattern{
		Initial_scale:       1,
		Intial_minRateKbits: 20400,
	}
	err := db_drp.ParseDRP(drp.NewDataRatePatternFileProvider(GOOD_PATH))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	counter := 0
	for {
		rate, err := db_drp.Next()
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
	db_drp := datatypes.DB_data_rate_pattern{
		Initial_scale:       1,
		Intial_minRateKbits: 20400,
	}
	err := db_drp.ParseDRP(drp.NewDataRatePatternFileProvider(SAW_PATH))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	if db_drp.GetHashStr() != "acd87822fa43d98efc9b854884336ff3" {
		t.Fatalf("Hash is incorrect. Is: '%s', should be %s", db_drp.GetHashStr(), "acd87822fa43d98efc9b854884336ff3")
	}
}

func TestDrpHashWithChangesToScaleMinLoop(t *testing.T) {
	db_drps := []datatypes.DB_data_rate_pattern{
		{
			Initial_scale:       1,
			Intial_minRateKbits: 20400,
		},
		{
			Initial_scale:       2,
			Intial_minRateKbits: 20400,
		},
		{
			Initial_scale:       1,
			Intial_minRateKbits: 50400,
		},
		{
			Initial_scale:       3,
			Intial_minRateKbits: 120400,
		},
	}
	for _, v := range db_drps {
		err := v.ParseDRP(drp.NewDataRatePatternFileProvider(SAW_PATH))
		if err != nil {
			t.Fatalf("Loaded valid drp, got an error: %s", err)
		}
		t.Log(v.GetStats())
		if v.GetHashStr() != "acd87822fa43d98efc9b854884336ff3" {
			t.Fatalf("Hash is incorrect. Is: '%s', should be %s", v.GetHashStr(), "acd87822fa43d98efc9b854884336ff3")
		}
	}

}

func TestBroken(t *testing.T) {

	files, err := os.ReadDir(paths.TESTDATA_DRP())
	if err != nil {
		t.Skipf("While reading testdata: %s", err)
	}
	for _, v := range files {
		if !strings.HasPrefix(v.Name(), "broken_") {
			continue
		}

		path := filepath.Join(paths.TESTDATA_DRP(), v.Name())
		db_drp := datatypes.DB_data_rate_pattern{
			Initial_scale:       1,
			Intial_minRateKbits: 20400,
		}
		err := db_drp.ParseDRP(drp.NewDataRatePatternFileProvider(filepath.Join(paths.TESTDATA_DRP(), v.Name())))
		if err == nil {
			t.Fatalf("Loaded Invalid Pattern '%s' -> %+v", path, db_drp)
		}

	}

}

func TestDrpValueNoLoop(t *testing.T) {
	var expected = []float64{10000, 20000, 30000, 40000, 50000, 60000, 70000, 80000, 90000, 100000}
	drp_db := datatypes.DB_data_rate_pattern{
		Initial_scale:       1,
		Intial_minRateKbits: 600,
	}
	err := drp_db.ParseDRP(drp.NewDataRatePatternFileProvider(SAW_PATH))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	counter := 0
	for {
		rate, err := drp_db.Next()
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
	db_drp := datatypes.DB_data_rate_pattern{
		Initial_scale:       1,
		Intial_minRateKbits: 600,
	}
	err := db_drp.ParseDRP(drp.NewDataRatePatternFileProvider(SAW_PATH))
	if err != nil {
		t.Fatalf("Loaded valid drp, got an error: %s", err)
	}
	db_drp.SetLooping(true)
	counter := 0
	for {
		rate, err := db_drp.Next()
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
