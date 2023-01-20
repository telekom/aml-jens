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

func TestIterator_saw(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")
	drp, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	iter := drp.Iterator()
	pos := 0
	for v, err := iter.Next(); err == nil; v, err = iter.Next() {
		t.Log(pos, v)
		if ExpectationSaw[pos] != v {
			t.Fatalf("Got unexpected value '%f' @ %d, (!= %f)", v, pos, ExpectationSaw[pos])
		}
		if pos > 10 {
			t.Fatal("No Iterator Exception was thrown")
		}
		pos += 1
	}
	if pos == 0 {
		t.Fatal("Nothing was iterated")
	}
}
func TestIteratorValue_saw(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")
	drp, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	iter := drp.Iterator()
	if iter.Value() != 10000 {
		t.Fatalf("Value before first call to Next() is %f", iter.Value())
	}
	pos := 0
	for v, err := iter.Next(); err == nil; v, err = iter.Next() {
		v_got := iter.Value()
		t.Log(v_got)
		if v_got != v {
			t.Fatalf("Got unexpected Value():'%f'!= %f)", iter.Value(), v)
		}
		pos++
		if pos > 11 {
			t.Fatal("Next() does not return error")
		}
	}
	if pos < 10 {
		t.Fatal("Not all entries in drp were played")
	}
}

func TestIterator_saw_loop(t *testing.T) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")
	drp, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		t.Fatal(err)
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
		100000,
		90000,
		80000,
		70000,
		60000,
		50000,
		40000,
		30000,
		20000,
		10000,
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
	iter := drp.Iterator()
	iter.SetLooping(true)
	pos := 0
	for v, err := iter.Next(); err == nil; v, err = iter.Next() {
		if pos < 30 && expected[pos] != v {
			t.Fatalf("Got unexpected value '%f' @ %d, (!= %f)", v, pos, expected[pos])
		}
		if pos > 103 {
			return //All good!
		}
		if iter.Value() != v {
			t.Fatalf("Value() %f != %f Next()", iter.Value(), v)
		}
		pos++
	}
	if pos == 0 {
		t.Fatal("Nothing was iterated")
	}
}

var result float64

func BenchmarkDRPIter(b *testing.B) {
	var path = filepath.Join(paths.TESTDATA_DRP(), "drp_3_valleys.csv")
	drp, err := NewDataRatePatternFileProvider(path).Provide(0, 0)
	if err != nil {
		b.Skip()
	}
	iter := drp.Iterator()
	iter.SetLooping(true)
	var pos int64 = 0
	var add float64
	for v, err := iter.Next(); err == nil; _, err = iter.Next() {
		pos++
		add += v
		if pos == 0xffffffff {
			break
		}
	}
	result = float64(pos) + add
}
