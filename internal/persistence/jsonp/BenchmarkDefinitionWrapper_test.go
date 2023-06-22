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
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/persistence/jsonp"
	"github.com/telekom/aml-jens/internal/util/utiltest"

	"github.com/spf13/viper"
)

func TestBenchmarkDefinitionWrapperJsonMarshall(t *testing.T) {
	viper.AddConfigPath(utiltest.TEST_CONFIG_PATH)
	const EXPECTED = "ff6a86c3b26621a1789b63311631ba73"
	data, err := jsonp.NewBenchmarkDefinitionWrapper(jsonp.BenchmarkDefinition{})
	if err != nil {
		t.Fatal(err)
	}

	txtbin, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	txt := string(txtbin)
	utiltest.InJsonOutput(t, txt, "Hash")
	utiltest.InJsonOutput(t, txt, "Inner")
	if data.HashStr != EXPECTED {
		t.Fatalf("Hash of new BenchmarkDefinitionWrapper does not match set default\n%s != %s", data.HashStr, EXPECTED)
	}
	if viper.ConfigFileUsed() != "" && !strings.HasSuffix(viper.ConfigFileUsed(), "/testdata/config/config.toml") {
		t.Fatalf("Loaded wrong config: %s", viper.ConfigFileUsed())
	}
}

func TestBenchmarkDefinitionWrapperJsonUnMarshall(t *testing.T) {
	tc := jsonp.NewDrPlayTrafficControlConfig(1, 2, 12, false, false)
	drp := jsonp.NewDrPlayDataRateConfig(10, 10, 1, 100)
	drplayset := jsonp.NewDrplaySetting(100, 1.0, 500, 11000)
	drplayset.TC = &tc
	payload, err := jsonp.NewBenchmarkDefinitionWrapper(
		jsonp.BenchmarkDefinition{
			Name:                      "TestingName001",
			Max_application_bitrate:   0xffffffff,
			MaxBitrateEstimationTimeS: 20,
			DrplaySetting:             drplayset,
			Patterns: []jsonp.BenchmarkPattern{
				{Path: filepath.Join(paths.TESTDATA_DRP(), "drp_3valleys.csv"), Setting: jsonp.DrplaySetting{
					TC:  &tc,
					DRP: &drp,
				}},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := payload.Validate(); err != nil {
		t.Fatalf("Correct NewBenchmarkDefinitionWrapper, would not validate: %s", err)
	}
	if marshalled, err := json.Marshal(payload); err != nil {
		t.Fatal(err)
	} else {
		data2 := &jsonp.BenchmarkDefinitionWrapper{}
		err = json.Unmarshal(marshalled, data2)
		if err != nil {
			t.Fatal(err)
		}
		if err := data2.Validate(); err != nil {
			t.Fatalf("Loaded NewBenchmarkDefinitionWrapper, would not validate: %s", err)
		}
		if payload.HashStr != data2.HashStr {
			t.Fatal("Hashed do not match")
		}
		if !reflect.DeepEqual(payload.Inner.DrplaySetting, data2.Inner.DrplaySetting) {
			t.Fatal("Inner DrplaySetting does not match")
		}
		if !reflect.DeepEqual(payload.Inner.Patterns, data2.Inner.Patterns) {
			t.Fatal("Inner DrplaySetting does not match")
		}
	}

}
