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
	"jens/drcommon/persistence/jsonp"
	"testing"
)

func TestNewBenchmarkPattern(t *testing.T) {
	PATH := "./testdata/drp/drp_3valleys.csv"
	data := jsonp.NewBenchmarkPattern(PATH, jsonp.NewDrplaySetting(100, 1.0, 500))
	if err := data.Validate(); err != nil {
		t.Fatalf("Validation failed: %s", err)
	}
	if data.Path != PATH {
		t.Fatal("Path not set correctly in NewBenchmarkPattern")
	}

	if data.Setting.TC != nil {
		t.Fatal("TC schould be nil, because nothing was explicitly set")
	}
	if *data.Setting.DRP.Frequency != 100 {
		t.Fatal("DrplaySetting was not set correctly in Pattern")
	}

}
