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

package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

func preTest(t *testing.T, args []string) func() {
	old := os.Args
	os.Args = make([]string, 1, len(args)+1)
	os.Args[0] = "testexecutables/drplay"
	os.Args = append(os.Args, args...)
	return func() {
		os.Args = old
	}
}
func TestArgParse(t *testing.T) {
	viper.GetViper().AddConfigPath("../../test/testdata/config/")
	//Sadly only able to test all, due to flag package
	//not having a reset function
	possibleArgs := [][]string{
		{"-csv"},
		{"-loop"},
		{"-freq", "36"},
		{"-dev", "eno1"},
		{"-pattern", "/tmp/drp_3valleys.csv"},
		//{"-psql"},
		{"-scale", "0.36"},
		{"-tag", "aTestTag36"},
	}
	test := [][]int{
		{0, 1, 2, 3, 4, 5, 6},
	}
	for _, tests := range test {
		args := make([]string, 0, len(tests))
		for _, v := range tests {
			args = append(args, possibleArgs[v]...)
		}
		post := preTest(t, args)
		config.PlayCfg().A_Session = &datatypes.DB_session{
			ParentBenchmark: &datatypes.DB_benchmark{
				Name:          "Not a part of a benchmark",
				Tag:           "N/A",
				PrintToStdOut: true,
				CsvOuptut:     false,
			},
			Time:     uint64(time.Now().UnixMilli()),
			ChildDRP: &datatypes.DB_data_rate_pattern{},
		}
		ArgParse()
		validate(t, tests, config.PlayCfg().A_Session, post)
		post()
	}

}
func validate(t *testing.T, set []int, sess *datatypes.DB_session, cleanup func()) {
	caseA := "Argparse: T(%d): %s was set, but NOT set in cfg"
	caseB := "Argparse: T(%d): %s was NOT set, but set in cfg"
	if !sess.ParentBenchmark.CsvOuptut {
		if contains(set, 0) {
			cleanup()
			t.Fatalf(caseA, 0, "-csv")
		} else {
			cleanup()
			t.Fatalf(caseB, 0, "-csv")
		}
	}
	if !sess.ChildDRP.Loop {
		if contains(set, 1) {
			cleanup()
			t.Fatalf(caseA, 1, "-loop")
		} else {
			cleanup()
			t.Fatalf(caseB, 1, "-loop")
		}
	}

	if sess.ChildDRP.Freq != 36 {
		if contains(set, 2) {
			cleanup()
			t.Fatalf(caseA, 2, "-frequency=36")
		} else {
			cleanup()
			t.Fatalf(caseB, 2, fmt.Sprintf("-frequency=%d", sess.ChildDRP.Freq))

		}
	}

	if sess.Dev != "eno1" {
		if contains(set, 3) {
			cleanup()
			t.Fatalf(caseA, 3, "-dev=eno1")
		} else {
			cleanup()
			t.Fatalf(caseB, 3, "-dev"+sess.Dev)
		}

	}

	if sess.ChildDRP.Name != "drp_3valleys.csv" {
		if contains(set, 4) {
			cleanup()
			t.Fatalf(caseA, 4, "-pattern='/tmp/drp_3valleys.csv'")
		} else {
			cleanup()
			t.Fatalf(caseB, 4, fmt.Sprintf("-pattern='%s'", sess.ChildDRP.Name))
		}

	}
	/*
		if err, pers := persistence.GetPersistence(); err != nil && reflect.TypeOf(pers) == reflect.TypeOf(psql.DataBase{}) {
			if contains(set, 5) {
				cleanup()
				t.Fatalf(caseA, 5, "-psql")
			} else {
				cleanup()
				t.Fatalf(caseB, 5, "-psql")
			}

		}
	*/

	if sess.ChildDRP.Scale != 0.36 {
		if contains(set, 5) {
			cleanup()
			t.Fatalf(caseA, 5, "-scale=0.36")
		} else {
			cleanup()
			t.Fatalf(caseB, 5, fmt.Sprintf("-scale=%f", sess.ChildDRP.Scale))
		}

	}
	if sess.Name != "aTestTag36" {
		if contains(set, 6) {
			cleanup()
			t.Fatalf(caseA, 6, "-tag='aTestTag36'")
		} else {
			cleanup()
			t.Fatalf(caseB, 6, fmt.Sprintf("-tag='%s'", sess.Name))
		}

	}

}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
