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
	"flag"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/persistence/mock"
	"github.com/telekom/aml-jens/internal/persistence/psql"
	"github.com/telekom/aml-jens/pkg/drp"
	drplay "github.com/telekom/aml-jens/pkg/drp_player"
)

var DEBUG, INFO, FATAL = logging.GetLogger()

func ArgParse() (err error) {
	result := config.PlayCfg().A_Session
	var looping bool
	// parse parameters
	flag.StringVar(
		&(result.Dev),
		"dev",
		"",
		"nic to play data rate pattern on, default 'lo'")

	pattern_path := flag.String(
		"pattern",
		"/etc/jens-cli/drp_3valleys.csv",
		"csv file for data rate pattern (seperator enter, values in kbits)")

	flag.IntVar(
		&result.ChildDRP.Freq,
		"freq",
		10,
		"number of samples per second to play [1 ... 100], default 10")

	flag.StringVar(
		&result.Name,
		"tag",
		time.Now().Format("2006.01.02 15:04:05"),
		"tag of this measure session")
	flag.BoolVar(
		&looping,
		"loop",
		false,
		"defines if data rate pattern player should run in an endless loop")
	flag.BoolVar(
		&result.ParentBenchmark.CsvOuptut,
		"csv",
		false,
		"output measure records to csv file")

	flag.Float64Var(
		&result.ChildDRP.Scale,
		"scale",
		1,
		"defines a scale factor which will be used to multiply the drp;  must be greater 0.1")

	postgresPtr := flag.Bool(
		"psql",
		false,
		"output measure records to configured postgresql db")

	flag.BoolVar(
		&result.ChildDRP.Nomeasure,
		"nomeasure",
		false,
		"only play drp, no queue measures are recorded")

	flag.Parse()
	if result.Dev == "" {
		logging.FlagParseExit("Flag: 'dev' was not set")
	}
	if *postgresPtr {
		err := persistence.SetPersistenceTo(&psql.DataBase{}, &config.PlayCfg().Psql)
		if err != nil {
			return err
		}
	} else {
		err := persistence.SetPersistenceTo(&mock.Database{}, &datatypes.Login{})
		if err != nil {
			return err
		}
	}
	err = result.ChildDRP.ParseDRP(drp.NewDataRatePatternFileProvider(*pattern_path))
	result.ChildDRP.SetLooping(looping)
	return err
}

func main() {
	logging.InitLogger(assets.NAME_DRPLAY)
	logging.EnableDebug()

	if err := ArgParse(); err != nil {
		FATAL.Println("Error during Argparse")
		FATAL.Exit(err)
	}
	session := config.PlayCfg().A_Session
	db, err := persistence.GetPersistence()
	if err != nil {
		FATAL.Exit(err)
	}
	err = (*db).Persist(session)
	if err != nil {
		FATAL.Exit(err)
	}
	drplay.StartDrpPlayer(session)
	if err := (*db).Close(); err != nil {
		FATAL.Exit(err)
	}
}
