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

package drbenchmark

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/persistence/jsonp"
	"github.com/telekom/aml-jens/internal/persistence/mock"
	"github.com/telekom/aml-jens/internal/util/utiltest"

	"github.com/spf13/viper"
)

// readability types
type TCcfg = jsonp.DrPlayTrafficControlConfig

type DRPcfg = jsonp.DrPlayDataRateConfig

// internal housekeeping
type TestString string

const (
	DRP_NAME = "drp_3valleys.csv"
	DRP_PATH = "./testdata/drp/" + DRP_NAME
)

var BENCHMARK_PATH = filepath.Join(paths.TESTDATA(), "benchmark") + "/%s.json"

const (
	T_NoIndividualSet    TestString = "NoIndividualSettingsSet"
	T_FallbackConfig     TestString = "PartiallySetFallbackCfg"
	T_CombinedTests      TestString = "CombinedTests"
	T_InvalidTopHash     TestString = "Broken_InvalidTopHash"
	T_InvalidPatternHash TestString = "Broken_InvalidPatternHash"
	T_AllIndividualsSet  TestString = "AllIndividualsSet"
	T_NoPatterns         TestString = "Broken_NoPatterns"
	T_NoPatternsEmptyObj TestString = "Broken_NoPatterns_EmptyObj"
	T_EmptyFile          TestString = "Broken_EmptyFile"
	T_EmptyFileJSONObj   TestString = "Broken_EmptyJsonObj"
	T_NoInner            TestString = "Broken_NoInner"
)

type oint32 struct {
	v int32
	o string
}
type oint struct {
	v int
	o string
}
type ofloat64 struct {
	v float64
	o string
}
type obool struct {
	v bool
	o string
}

type ostring struct {
	v string
	o string
}

type drpv struct {
	WarmupTimeMs oint32
	Freq         oint
	Scale        ofloat64
	MinRateKbits ofloat64
	Name         ostring
}
type sessv struct {
	Markfree            oint32
	Markfull            oint32
	Extralatency        oint32
	L4sEnablePreMarking obool
	WarmupTimeMs        oint
	SignalDrpStart      obool
	Queuesizepackets    oint32
}

func TestMain(m *testing.M) {
	err := exec.Command("cp", "-r", paths.TESTDATA(), "./testdata").Run()
	if err != nil {
		println(err)
		os.Exit(-1)
	}
	m.Run()
	err = exec.Command("rm", "-rf", "./testdata").Run()
	if err != nil {
		println(err)
		os.Exit(2)
	}
}

func verifyDRP(t *testing.T, prefix string, got *datatypes.DB_data_rate_pattern, e drpv) {
	var log_msg = prefix + " DRP -> %s Got:%+v ; Expected:%+v (from %s)"
	if got.WarmupTimeMs != e.WarmupTimeMs.v {
		t.Fatalf(log_msg, "WarmupTimeMs", got.WarmupTimeMs, e.WarmupTimeMs.v, e.WarmupTimeMs.o)
	}
	if got.Freq != e.Freq.v {
		t.Fatalf(log_msg, "Freq", got.Freq, e.Freq.v, e.Freq.o)
	}
	if got.Initial_scale != e.Scale.v {
		t.Fatalf(log_msg, "Scale", got.Initial_scale, e.Scale.v, e.Scale.o)
	}
	if got.Intial_minRateKbits != e.MinRateKbits.v {
		t.Fatalf(log_msg, "MinRateKbits", got.Intial_minRateKbits, e.MinRateKbits.v, e.MinRateKbits.o)
	}
	if got.GetName() != e.Name.v {
		t.Fatalf(log_msg, "Name", got.GetName(), e.Name.v, e.Name.o)
	}
}
func verifySession(t *testing.T, prefix string, got *datatypes.DB_session, e sessv) {
	var log_msg = prefix + " Session -> %s Got:%+v ; Expected:%+v (from %s)"

	if got.Markfree != e.Markfree.v {
		t.Fatalf(log_msg, "Markfree", got.Markfree, e.Markfree.v, e.Markfree.o)
	}
	if got.Markfull != e.Markfull.v {
		t.Fatalf(log_msg, "Markfull", got.Markfull, e.Markfull.v, e.Markfull.o)
	}
	if got.ExtralatencyMs != e.Extralatency.v {
		t.Fatalf(log_msg, "Extralatency", got.ExtralatencyMs, e.Extralatency.v, e.Extralatency.o)
	}
	if got.L4sEnablePreMarking != e.L4sEnablePreMarking.v {
		t.Fatalf(log_msg, "L4sEnablePreMarking", got.L4sEnablePreMarking, e.L4sEnablePreMarking.v, e.L4sEnablePreMarking.o)
	}
	if got.SignalDrpStart != e.SignalDrpStart.v {
		t.Fatalf(log_msg, "SignalDrpStart", got.SignalDrpStart, e.SignalDrpStart.v, e.SignalDrpStart.o)
	}
	if got.Queuesizepackets != e.Queuesizepackets.v {
		t.Fatalf(log_msg, "Queuesizepackets", got.Queuesizepackets, e.Queuesizepackets.v, e.Queuesizepackets.o)
	}
}

func TestReadBenchmarkFromFileConfigFallback(t *testing.T) {
	viper.AddConfigPath(utiltest.TEST_CONFIG_PATH)
	persistence.SetPersistenceTo(&mock.Database{}, &datatypes.Login{})
	bm, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_FallbackConfig))
	if err != nil {
		t.Fatalf("Could not load valid Benchmark: %s", err)
	}
	if !strings.HasSuffix(viper.ConfigFileUsed(), "/testdata/config/config.toml") {
		t.Fatalf("Loaded wrong config: %s", viper.ConfigFileUsed())
	}
	got := bm.Sessions[0].ChildDRP
	verifyDRP(t, string(T_FallbackConfig), got, drpv{
		WarmupTimeMs: oint32{9999, "cfg"},
		Freq:         oint{76, "json-pattern"},
		Scale:        ofloat64{0.76, "json-pattern"},
		MinRateKbits: ofloat64{7600, "json-inner"},
		Name:         ostring{"drp_3valleys.csv", "json-pattern"},
	})
	verifySession(t, string(T_FallbackConfig), bm.Sessions[0], sessv{
		Markfree:            oint32{7, "json-pattern"},
		Markfull:            oint32{76, "json-inner"},
		Extralatency:        oint32{10, "cfg"},
		Queuesizepackets:    oint32{10000, "cfg"},
		L4sEnablePreMarking: obool{true, "cfg"},
		SignalDrpStart:      obool{true, "cfg"},
	})
}

func TestReadBenchmarkFromFileAllSettingsInner(t *testing.T) {
	viper.AddConfigPath(utiltest.TEST_CONFIG_PATH)
	persistence.SetPersistenceTo(&mock.Database{}, nil)
	bm, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_NoIndividualSet))
	if err != nil {
		t.Fatal(err)
	}
	bm.Tag = "TESTING"
	for _, v := range bm.Sessions {
		inter, err := net.InterfaceByIndex(1)
		if err != nil {
			t.Skipf("Skipping due to %s", err)
		}
		v.Dev = inter.Name
	}
	if err != nil {
		t.Fatalf("Could not load valid Benchmark: %s", err)
	}
	err = bm.Validate()
	if err != nil {
		t.Fatalf("Could not Validate benchmark: %s", err)
	}
	for _, v := range bm.Sessions {
		verifyDRP(t, string(T_NoIndividualSet), v.ChildDRP, drpv{
			WarmupTimeMs: oint32{7600, "json-inner"},
			Freq:         oint{76, "json-inner"},
			Scale:        ofloat64{0.76, "json-inner"},
			MinRateKbits: ofloat64{7600, "json-inner"},
			Name:         ostring{"drp_3valleys.csv", "json-pattern"},
		})
		verifySession(t, string(T_NoIndividualSet), bm.Sessions[0], sessv{
			Markfree:            oint32{7, "json-inner"},
			Markfull:            oint32{76, "json-inner"},
			Extralatency:        oint32{13, "json-inner"},
			Queuesizepackets:    oint32{10001, "json-inner"},
			L4sEnablePreMarking: obool{false, "json-inner"},
			SignalDrpStart:      obool{true, "json-inner"},
		})
	}
}
func TestReadBenchmarkFromFileAllSettingsPattern(t *testing.T) {
	viper.AddConfigPath(utiltest.TEST_CONFIG_PATH)
	persistence.SetPersistenceTo(&mock.Database{}, nil)
	bm, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_AllIndividualsSet))
	if err != nil {
		t.Fatal(err)
	}
	bm.Tag = "TESTING"
	for _, v := range bm.Sessions {
		inter, err := net.InterfaceByIndex(1)
		if err != nil {
			t.Skipf("Skipping due to %s", err)
		}
		v.Dev = inter.Name
	}
	if err != nil {
		t.Fatalf("Could not load valid Benchmark: %s", err)
	}
	err = bm.Validate()
	if err != nil {
		t.Fatalf("Could not Validate benchmark: %s", err)
	}
	for _, v := range bm.Sessions {
		verifyDRP(t, string(T_AllIndividualsSet), v.ChildDRP, drpv{
			WarmupTimeMs: oint32{42, "json-pattern"},
			Freq:         oint{42, "json-pattern"},
			Scale:        ofloat64{0.42, "json-pattern"},
			MinRateKbits: ofloat64{4200, "json-pattern"},
			Name:         ostring{"drp_3valleys.csv", "json-pattern"},
		})
		verifySession(t, string(T_AllIndividualsSet), bm.Sessions[0], sessv{
			Markfree:            oint32{2, "json-pattern"},
			Markfull:            oint32{4, "json-pattern"},
			Extralatency:        oint32{2, "json-inner"},
			Queuesizepackets:    oint32{10003, "json-pattern"},
			L4sEnablePreMarking: obool{true, "json-pattern"},
			SignalDrpStart:      obool{true, "json-pattern"},
		})
	}
}

func TestReadBenchmarkFromFile(t *testing.T) {
	persistence.SetPersistenceTo(&mock.Database{}, &datatypes.Login{})

	bm, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_CombinedTests))
	if err != nil {
		t.Fatal(err)
	}
	bm.Tag = "TESTING"
	for i, v := range bm.Sessions {
		if v.ChildDRP.GetName() != DRP_NAME {
			t.Fatalf("Name for pattern %d is read incorrectly. (is: %s, should be %s)", i, v.ChildDRP.GetName(), DRP_NAME)
		}
	}
	expectations := [][3]interface{}{
		/*0*/ {
			"All in json-pattern",
			drpv{
				WarmupTimeMs: oint32{v: 4, o: "json-pattern"},
				Freq:         oint{v: 1, o: "json-pattern"},
				Scale:        ofloat64{v: 0.2222, o: "json-pattern"},
				MinRateKbits: ofloat64{v: 3, o: "json-pattern"},
				Name:         ostring{"drp_3valleys.csv", "json-pattern"},
			}, sessv{
				Markfree:            oint32{v: 1, o: "json-pattern"},
				Markfull:            oint32{v: 2, o: "json-pattern"},
				Extralatency:        oint32{0, "json-inner"},
				Queuesizepackets:    oint32{10007, "json-pattern"},
				L4sEnablePreMarking: obool{v: false, o: "json-pattern"},
				SignalDrpStart:      obool{v: false, o: "json-pattern"},
			}},
		/*1*/ {
			"All in json-inner; some in tc",
			drpv{
				WarmupTimeMs: oint32{v: 5000, o: "json-inner"},
				Freq:         oint{v: 87, o: "json-inner"},
				Scale:        ofloat64{v: 0.87654, o: "json-inner"},
				MinRateKbits: ofloat64{v: 5000, o: "json-inner"},
				Name:         ostring{"drp_3valleys.csv", "json-pattern"},
			}, sessv{
				Markfree:            oint32{v: 1, o: "json-pattern"},
				Markfull:            oint32{v: 2, o: "json-pattern"},
				Queuesizepackets:    oint32{10000, "json-inner"},
				Extralatency:        oint32{10, "json-inner"},
				L4sEnablePreMarking: obool{v: false, o: "json-inner"},
				SignalDrpStart:      obool{v: true, o: "json-inner"},
			}},
		/*2*/ {
			"tc in inner; drp in pattern",
			drpv{
				WarmupTimeMs: oint32{v: 4, o: "json-pattern"},
				Freq:         oint{v: 1, o: "json-pattern"},
				Scale:        ofloat64{v: 0.2, o: "json-pattern"},
				MinRateKbits: ofloat64{v: 3, o: "json-pattern"},
				Name:         ostring{"drp_3valleys.csv", "json-pattern"},
			}, sessv{
				Markfree:            oint32{v: 7, o: "json-inner"},
				Markfull:            oint32{v: 77, o: "json-inner"},
				Queuesizepackets:    oint32{10000, "json-inner"},
				Extralatency:        oint32{10, "json-inner"},
				L4sEnablePreMarking: obool{v: false, o: "json-inner"},
				SignalDrpStart:      obool{v: true, o: "json-inner"},
			}},
		/*3*/ {
			"Empty objects Pattern",
			drpv{
				WarmupTimeMs: oint32{v: 5000, o: "json-inner"},
				Freq:         oint{v: 87, o: "json-inner"},
				Scale:        ofloat64{v: 0.87654, o: "json-inner"},
				MinRateKbits: ofloat64{v: 5000, o: "json-inner"},
				Name:         ostring{"drp_3valleys.csv", "json-pattern"},
			}, sessv{
				Markfree:            oint32{v: 7, o: "json-inner"},
				Markfull:            oint32{v: 77, o: "json-inner"},
				Queuesizepackets:    oint32{10000, "json-inner"},
				Extralatency:        oint32{10, "json-inner"},
				L4sEnablePreMarking: obool{v: false, o: "json-inner"},
				SignalDrpStart:      obool{v: true, o: "json-inner"},
			}},
		/*4*/ {
			"Empty Setting object Pattern",
			drpv{
				WarmupTimeMs: oint32{v: 5000, o: "json-inner"},
				Freq:         oint{v: 87, o: "json-inner"},
				Scale:        ofloat64{v: 0.87654, o: "json-inner"},
				MinRateKbits: ofloat64{v: 5000, o: "json-inner"},
				Name:         ostring{"drp_3valleys.csv", "json-pattern"},
			}, sessv{
				Markfree:            oint32{v: 7, o: "json-inner"},
				Markfull:            oint32{v: 77, o: "json-inner"},
				Queuesizepackets:    oint32{10000, "json-inner"},
				Extralatency:        oint32{10, "json-inner"},
				L4sEnablePreMarking: obool{v: false, o: "json-inner"},
				SignalDrpStart:      obool{v: true, o: "json-inner"},
			}},
		/*5*/ {
			"Missing Setting object Pattren",
			drpv{
				WarmupTimeMs: oint32{v: 5000, o: "json-inner"},
				Freq:         oint{v: 87, o: "json-inner"},
				Scale:        ofloat64{v: 0.87654, o: "json-inner"},
				MinRateKbits: ofloat64{v: 5000, o: "json-inner"},
				Name:         ostring{"drp_3valleys.csv", "json-pattern"},
			}, sessv{
				Markfree:            oint32{v: 7, o: "json-inner"},
				Markfull:            oint32{v: 77, o: "json-inner"},
				Queuesizepackets:    oint32{10000, "json-inner"},
				Extralatency:        oint32{10, "json-inner"},
				L4sEnablePreMarking: obool{v: false, o: "json-inner"},
				SignalDrpStart:      obool{v: true, o: "json-inner"},
			}},
	}
	for i, v := range expectations {
		desc := v[0].(string)
		drpE := v[1].(drpv)
		sesE := v[2].(sessv)
		t.Logf("[%d] %s", i, desc)
		verifyDRP(t, fmt.Sprintf("%s (%s)", T_CombinedTests, desc), bm.Sessions[i].ChildDRP, drpE)
		verifySession(t, fmt.Sprintf("%s (%s)", T_CombinedTests, desc), bm.Sessions[i], sesE)
	}
}

func TestInvalidTopLevelHash(t *testing.T) {

	config.PlayCfg().A_Session.ChildDRP.Freq = 10
	config.PlayCfg().A_Session.ChildDRP.Initial_scale = 0.10
	bm, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_InvalidTopHash))
	if err == nil {
		t.Logf("%+v", bm)
		log.Fatalf("Benchmark with incorrect hash was loaded")
	}
	if !strings.Contains(err.Error(), "TopLevelHash") {
		t.Logf("%+v", bm)
		log.Fatalf("Benchmark was rejected due to wrong reason: %s", err)
	}
}

func TestInvalidPatternHash(t *testing.T) {
	config.PlayCfg().A_Session.ChildDRP.Freq = 10
	config.PlayCfg().A_Session.ChildDRP.Initial_scale = 0.10

	bm, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_InvalidPatternHash))
	if err == nil {
		t.Logf("%+v", bm)
		t.Fatalf("Benchmark with incorrect hash was loaded")
	}
	if !strings.Contains(err.Error(), "DRP-Hash") {
		t.Logf("%+v", bm)
		t.Fatalf("Benchmark was rejected due to wrong reason: %s", err)
	}
}

func TestBrokenBenchmarkNoPatterns(t *testing.T) {
	viper.AddConfigPath(utiltest.TEST_CONFIG_PATH)
	persistence.SetPersistenceTo(&mock.Database{}, nil)
	_, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_NoPatterns))
	if err == nil {
		t.Fatal("Loaded benchmark without patterns")
	}
}
func TestBrokenBenchmarkNoPatternsEmptyObj(t *testing.T) {
	viper.AddConfigPath(utiltest.TEST_CONFIG_PATH)
	persistence.SetPersistenceTo(&mock.Database{}, nil)
	_, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_NoPatternsEmptyObj))
	if err == nil {
		t.Fatal("Loaded benchmark without patterns (empty obj)")
	}
}
func TestBrokenBenchmarkEmptyFile(t *testing.T) {
	_, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_EmptyFile))
	if err == nil {
		t.Fatal("Loaded broken benchmark (T_EmptyFile)")
	}
}
func TestBrokenBenchmarkEmptyFileJSONObj(t *testing.T) {
	_, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_EmptyFileJSONObj))
	if err == nil {
		t.Fatal("Loaded broken benchmark (T_EmptyFileJSONObj)")
	}
}
func TestBrokenBenchmarkNoInner(t *testing.T) {
	_, err := jsonp.LoadDB_benchmarkFromJson(fmt.Sprintf(BENCHMARK_PATH, T_NoInner))
	if err == nil {
		t.Fatal("Loaded broken benchmark (T_NoInner)")
	}
}
