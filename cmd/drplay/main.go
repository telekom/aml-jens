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
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

func ArgParse() error {
	result := config.PlayCfg().A_Session
	var err error
	var looping bool
	var frequency int
	var bandwidthInUnit string
	var scale float64
	var bandwidthkbits int = 0
	var pattern_path string
	// parse parameters
	version := flag.Bool("v", false, "prints build version")
	flag.StringVar(
		&(result.Dev),
		"dev",
		"",
		"nic to play data rate pattern on, default 'lo'")
	flag.StringVar(
		&result.Name,
		"tag",
		time.Now().Format("2006.01.02 15:04:05"),
		"tag of this measure session")
	flag.StringVar(
		&pattern_path,
		"pattern",
		"",
		"csv file for data rate pattern (seperator enter, values in kbits), e.g. /etc/jens-cli/drp_3valleys.csv")

	flag.IntVar(
		&frequency,
		"freq",
		10,
		"number of samples per second to play [1 ... 100], default 10")
	flag.BoolVar(
		&looping,
		"loop",
		false,
		"defines if data rate pattern player should run in an endless loop")
	flag.Float64Var(
		&scale,
		"scale",
		1,
		"defines a scale factor which will be used to multiply the drp;  must be greater 0.1")

	flag.StringVar(&bandwidthInUnit,
		"bandwidth",
		"",
		"total bandwidth in unit m(mbits) or k(kbits)")

	flag.BoolVar(
		&result.ParentBenchmark.CsvOuptut,
		"csv",
		false,
		"output measure records to csv file")

	postgresPtr := flag.Bool(
		"psql",
		false,
		"output measure records to configured postgresql db")

	flag.BoolVar(
		&result.Nomeasure,
		"nomeasure",
		false,
		"only play drp, no queue measures are recorded")

	flag.Parse()
	cfg := config.PlayCfg()
	if *version {
		fmt.Printf("Version      : %s\n", assets.VERSION)
		fmt.Printf("Compiletime  : %s\n", assets.BUILD_TIME)
		os.Exit(0)
	}
	if result.Dev == "" {
		return fmt.Errorf("flag: 'dev' was not set")
	}
	result.ChildDRP.Initial_scale = scale
	result.ChildDRP.Freq = frequency
	cfg.A_MultiSession.Name = result.Name
	if bandwidthInUnit != "" {
		unit := strings.ToLower(bandwidthInUnit[len(bandwidthInUnit)-1:])
		bandwidthkbits, err = strconv.Atoi(bandwidthInUnit[:len(bandwidthInUnit)-1])
		if err != nil {
			logging.FlagParseExit("specified bandwidth is not a number")
		}
		if unit == "m" {
			bandwidthkbits *= 1000
		} else if unit != "k" {
			return fmt.Errorf("specified bandwidth must be in unit m(mbits) or k(kbits), e.g. 20000k")
		}
		config.PlayCfg().A_MultiSession.Bandwidthkbits = bandwidthkbits
		config.PlayCfg().A_MultiSession.DrpMode = false
	} else {
		//Pattern Mode
		if pattern_path == "" {
			pattern_path = "/etc/jens-cli/drp_3valleys.csv"
		}
		err := result.ChildDRP.ParseDRP(drp.NewDataRatePatternFileProvider(pattern_path))
		if err != nil {
			return fmt.Errorf("could not load pattern from path %s", pattern_path)
		}
		result.ChildDRP.SetLooping(looping)
	}

	if *postgresPtr {
		err := persistence.SetPersistenceTo(&psql.DataBase{}, &config.PlayCfg().Psql)
		if err != nil {
			WARN.Println(err)
			return fmt.Errorf("could not connect to the psql database")
		}
	} else {
		err := persistence.SetPersistenceTo(&mock.Database{}, &datatypes.Login{})
		if err != nil {
			return err
		}
	}
	return nil
}

func exithandler(player *drplay.DrpPlayer, exit chan uint8) {

	exit_handler := make(chan os.Signal)
	signal.Notify(exit_handler, syscall.SIGINT, syscall.SIGPIPE, syscall.SIGQUIT)
	go func() {
		select {
		case <-exit:
			return
		case sig := <-exit_handler:
			INFO.Printf("Received Signal: %d", sig)
			player.Exit()
		}
	}()
}

func exitHandler(player *drplay.DrpPlayer) {
	exit_handler := make(chan os.Signal)
	signal.Notify(exit_handler, syscall.SIGINT, syscall.SIGPIPE, syscall.SIGQUIT)
	go func() {
		sig := <-exit_handler
		// Run Cleanup
		INFO.Printf("Received Signal: %d\n", sig)
		player.Exit_clean()
		INFO.Printf("Resources resetted\n")
		os.Exit(1)
	}()
}

func main() {
	logging.InitLogger(assets.NAME_DRPLAY)
	INFO.Printf("===>Starting DrPlay @%s <===\n\n", time.Now().String())
	var player_has_ended = make(chan uint8)
	if err := ArgParse(); err != nil {
		logging.FlagParseExit(err.Error())
	}

	db, err := persistence.GetPersistence()
	if err != nil {
		FATAL.Println(err)
		os.Exit(4)
	}

	cfg := config.PlayCfg()
	player := drplay.NewDrpPlayer(cfg)

	//Todo should be done in caller, not callee
	logging.LinkExitFunction(func() uint8 {
		FATAL.Println("Logger experienced a fatal error")
		player.Exit()
		panic("A")
		//return 255
	}, 5000)
	exitHandler(player)

	err = player.Start()
	if err != nil {
		FATAL.Println(err)
		os.Exit(-1)
	}
	player.Wait()
	close(player_has_ended)
	if err := (*db).Close(); err != nil {
		FATAL.Println(err)
		os.Exit(2)
	}
	fmt.Println(strings.Join(assets.END_OF_DRPLAY[:], " "))
}
