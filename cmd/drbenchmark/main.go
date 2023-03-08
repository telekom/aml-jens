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
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	drbenchmark "github.com/telekom/aml-jens/cmd/drbenchmark/internal"
	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/persistence/jsonp"
	"github.com/telekom/aml-jens/internal/persistence/psql"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

func ArgParse() (*datatypes.DB_benchmark, error) {
	var dev string = ""
	var benchmark string = ""
	var tag string = ""
	version := flag.Bool("v", false, "prints build version")
	flag.StringVar(&dev, "dev", "",
		"nic to play data rate pattern on, default 'lo'")
	flag.StringVar(&benchmark, "benchmark", "/etc/jens-cli/benchmark_example.json",
		"JSON file containing a benchmark definition")
	flag.StringVar(&tag, "tag", "<interactive>",
		"name for the benchmark in DB.\nConvention: <algorithm> - L4S: <true/false>")
	flag.Parse()
	if *version {
		fmt.Printf("Version      : %s\n", assets.VERSION)
		fmt.Printf("Compiletime  : %s\n", assets.BUILD_TIME)
		os.Exit(0)
	}
	if len(flag.Args()) > 0 {
		logging.FlagParseExit("Unexpected Argument(s): '%v'", flag.Args())
	}
	if dev == "" {
		logging.FlagParseExit("Flag: 'dev' was not set")
	}
	var err error
	if tag == "<interactive>" {
		tag, err = askTag()
	}
	if err != nil {
		if v, ok := err.(*errortypes.UserInputError); ok {
			logging.FlagParseExit(v.Error())
		} else {
			INFO.Println(err)
			tag = "N/A - " + time.Now().Format("2006.01.02 15:04:05")
		}
	}
	res, err := jsonp.LoadDB_benchmarkFromJson(benchmark)
	if err != nil {
		return nil, fmt.Errorf("Could not load Benchmark from json: %w", err)
	}
	res.Tag = tag
	if _, err := net.InterfaceByName(dev); err != nil {
		return nil, fmt.Errorf("'%s' is not a recognized interface -> %v", dev, err)
	}
	for _, v := range res.Sessions {
		v.Dev = dev
	}
	return res, nil
}

func askTag() (string, error) {
	// get the FileInfo struct describing the standard input.
	fi, _ := os.Stdin.Stat()

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return "", errortypes.NewUserInputError("Can't use interactive mode")
	}
	fmt.Println("data is from terminal")
	// do things from data from terminal

	ConsoleReader := bufio.NewReader(os.Stdin)

	fmt.Print("Algorithm under test: ")
	alg, err := ConsoleReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	fmt.Print("L4s Enabled (Y/n)?: ")
	l4ss, err := ConsoleReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	l4ss = strings.TrimRight(l4ss, "\n")
	l4ss = strings.TrimRight(l4ss, " ")
	l4ss = strings.ToLower(l4ss)
	l4s_enabled := l4ss == "" || l4ss == "y" || l4ss == "j" || l4ss == "yes"
	return fmt.Sprintf("%s - L4S: %t", alg[:len(alg)-1], l4s_enabled), nil

}
func exithandler(bm *drbenchmark.Benchmark) chan uint8 {
	exit := make(chan uint8)
	exit_handler := make(chan os.Signal)
	signal.Notify(exit_handler, syscall.SIGINT, syscall.SIGPIPE, syscall.SIGQUIT)
	go func() {
		for {
			select {
			case <-exit:
				return
			case <-exit_handler:
				bm.SkipSession()
			}
		}
	}()
	return exit
}

func main() {
	logging.InitLogger(assets.NAME_DRBENCH)
	bm, err := ArgParse()
	if err != nil {
		FATAL.Println(err)
		return
	}
	if err := bm.Validate(); err != nil {
		FATAL.Println("Benchmark Validation failed")
		return
	}
	err = persistence.SetPersistenceTo(&psql.DataBase{}, &config.PlayCfg().Psql)
	if err != nil {
		FATAL.Println(err)
		return
	}
	benchmark := drbenchmark.New(bm)
	ex := exithandler(benchmark)
	defer close(ex)
	if err := benchmark.Play(); err != nil {
		FATAL.Println(err)
		os.Exit(-1)
	}
}
