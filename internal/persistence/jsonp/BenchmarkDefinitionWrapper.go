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

package jsonp

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
	"github.com/telekom/aml-jens/pkg/drp"
)

var DEBUG, INFO, FATAL = logging.GetLogger()

type BenchmarkDefinitionWrapper struct {
	hash    []byte
	HashStr string `json:"Hash"`
	Inner   BenchmarkDefinition
}

func NewBenchmarkDefinitionWrapper(def BenchmarkDefinition) (BenchmarkDefinitionWrapper, error) {
	b := BenchmarkDefinitionWrapper{
		Inner: def,
		hash:  make([]byte, 0, md5.Size),
	}
	return b, b.CheckOrCalcHash()
}
func LoadBenchmarkDefinitionFromJsonData(dat []byte) (*BenchmarkDefinition, string, error) {
	res := BenchmarkDefinitionWrapper{}
	dec := json.NewDecoder(bytes.NewReader(dat))
	dec.DisallowUnknownFields()

	err := dec.Decode(&res)
	if err != nil {
		return nil, "", err
	}
	//json.Unmarshal(dat, &res)
	err = res.CheckOrCalcHash()
	return &res.Inner, res.HashStr, err
}
func ReadTcValuesWithFallbacks(fb *datatypes.DB_session, tc ...*DrPlayTrafficControlConfig) (mark_free int32, mark_full int32, extralatency int32, l4spre bool, signalstart bool, queue_size int32) {
	mark_free = fb.Markfree
	mark_full = fb.Markfull
	extralatency = fb.ExtralatencyMs
	queue_size = fb.Queuesizepackets
	l4spre = fb.L4sEnablePreMarking
	signalstart = fb.SignalDrpStart

	mark_free_set := false
	mark_full_set := false
	extralatency_set := false
	queue_size_set := false
	l4spre_set := false
	signalstart_set := false

	for _, v := range tc {
		if v == nil {
			continue
		}
		if !mark_free_set && v.Markfree != nil {
			mark_free_set = true
			mark_free = *v.Markfree
		}
		if !mark_full_set && v.Markfull != nil {
			mark_full_set = true
			mark_full = *v.Markfull
		}
		if !extralatency_set && v.Extralatency != nil {
			extralatency_set = true
			extralatency = *v.Extralatency
		}
		if !queue_size_set && v.Queuesizepackets != nil {
			queue_size_set = true
			queue_size = int32(*v.Queuesizepackets)
		}
		if !l4spre_set && v.L4sEnablePreMarking != nil {
			l4spre_set = true
			l4spre = *v.L4sEnablePreMarking
		}
		if !signalstart_set && v.SignalDrpStart != nil {
			signalstart_set = true
			signalstart = *v.SignalDrpStart
		}
	}

	return mark_free, mark_full, extralatency, l4spre, signalstart, queue_size

}
func ReadDrpValuesWithFallbacks(fb *datatypes.DB_data_rate_pattern, drp ...*DrPlayDataRateConfig) (scale float64, freq int, minrate float64, warmup int32) {
	freq = fb.Freq
	scale = fb.Scale
	minrate = fb.MinRateKbits
	warmup = fb.WarmupTimeMs
	freq_set := false
	scale_set := false
	minrate_set := false
	warmup_set := false
	for _, v := range drp {
		if v == nil {
			continue
		}
		if !freq_set && v.Frequency != nil {
			freq = *v.Frequency
			freq_set = true
		}

		if !scale_set && v.Scale != nil {
			scale = *v.Scale
			scale_set = true
		}
		if !minrate_set && v.MinRateKbits != nil {
			minrate = *v.MinRateKbits
			minrate_set = true
		}
		if !warmup_set && v.WarmupBeforeDrpMs != nil {
			warmup = int32(*v.WarmupBeforeDrpMs)
			warmup_set = true
		}

	}
	return scale, freq, minrate, warmup
}

func LoadDB_benchmarkFromJson(path string) (*datatypes.DB_benchmark, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	defintion, hash, err := LoadBenchmarkDefinitionFromJsonData(data)
	if err != nil {
		return nil, err
	}
	if err := defintion.Validate(); err != nil {
		return nil, err
	}
	play_cfg := config.PlayCfg()
	benchmark := &datatypes.DB_benchmark{
		PrintToStdOut: false,
		Name:          defintion.Name,
		Tag:           config.BenchmarkCfg().A_Tag,
		Sessions:      make([]*datatypes.DB_session, len(defintion.Patterns)),
		Hash:          hash,
	}

	for i, v := range defintion.Patterns {
		s, f, m, w := ReadDrpValuesWithFallbacks(play_cfg.A_Session.ChildDRP, v.Setting.DRP, defintion.DrplaySetting.DRP)
		db_drp := datatypes.NewDB_data_rate_pattern()
		db_drp.SetLooping(false)
		db_drp.Freq = f
		db_drp.Scale = s
		db_drp.MinRateKbits = m
		db_drp.WarmupTimeMs = w

		db_drp.ParseDRP(drp.NewDataRatePatternFileProvider(v.Path))
		fe, fu, el, l4, ss, qs := ReadTcValuesWithFallbacks(play_cfg.A_Session, v.Setting.TC, defintion.DrplaySetting.TC)
		benchmark.Sessions[i] = &datatypes.DB_session{
			Markfree:            fe,
			Markfull:            fu,
			ExtralatencyMs:      el,
			L4sEnablePreMarking: l4,
			SignalDrpStart:      ss,
			Dev:                 config.BenchmarkCfg().A_Dev,
			ChildDRP:            db_drp,
			Name:                fmt.Sprintf("%s:%s (%d/%d)", benchmark.Name, benchmark.Tag, i+1, len(benchmark.Sessions)),
			Queuesizepackets:    qs,
		}
		benchmark.Sessions[i].SetParentBenchmark(benchmark)
	}

	return benchmark, nil
}

func (bmdw *BenchmarkDefinitionWrapper) CheckOrCalcHash() (err error) {
	if len(bmdw.hash) == 0 {
		bmdw.hash, err = bmdw.Inner.CalcMd5()
		if err != nil {
			return err
		}
	}
	generated_hashStr := hex.EncodeToString(bmdw.hash)
	if bmdw.HashStr == "" {
		bmdw.HashStr = generated_hashStr
		return nil
	}
	if bmdw.HashStr != generated_hashStr {
		return fmt.Errorf("TopLevelHashes dont match: %s != %s (set != calculated)", bmdw.HashStr, generated_hashStr)
	}
	return nil
}

func (wbcfg *BenchmarkDefinitionWrapper) HumanReadableSummary() string {
	return fmt.Sprintf("Benchmark: %s\n    Patterns: %d\n",
		wbcfg.Inner.Name, len(wbcfg.Inner.Patterns))
}
func (wbcfg *BenchmarkDefinitionWrapper) Validate() error {
	if len(wbcfg.hash) == 0 {
		INFO.Println("Benchmark has no hash-value")
		defer func() {
			wbcfg.CheckOrCalcHash()
		}()

	} else {
		calculated, err := wbcfg.Inner.CalcMd5()
		if err != nil {
			return err
		}
		if !util.ByteCompare(wbcfg.hash, calculated) {
			INFO.Printf("Benchmark invalid hash: '%s' != '%s'\n", wbcfg.HashStr, hex.EncodeToString(calculated))
			return fmt.Errorf("invalid Benchmark hash: is '%s', should be '%s'", hex.EncodeToString(calculated), wbcfg.HashStr)
		}
	}
	return wbcfg.Inner.Validate()
}
