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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/telekom/aml-jens/cmd/drshow/internal/modes/file"
	"github.com/telekom/aml-jens/cmd/drshow/internal/modes/folder"
	"github.com/telekom/aml-jens/cmd/drshow/internal/modes/pipe"
	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/logging"
)

type mode uint8

const (
	mode_pipe mode = iota
	mode_static
	mode_err
	mode_help
)

var _, INFO, FATAL = logging.GetLogger()

func parseArgs(res *string) (mode, error) {
	Nargs := len(os.Args) - 1
	INFO.Printf("Called with Args: %v - %d\n", os.Args, Nargs)
	if Nargs == 0 {
		return mode_pipe, nil
	}
	first_arg_is_help := helpstringMap[strings.ToUpper(os.Args[1])]
	if Nargs > 2 {
		return mode_err, errors.New("invalid Arguments. See man pages or drshow --help for more info")
	}

	if first_arg_is_help {
		*res = string(parseArg2ForHelp(os.Args[Nargs]))
		INFO.Println("Displaying help: " + *res)
		return mode_help, nil
	}
	*res = os.Args[Nargs]
	return mode_static, nil
}

func main() {
	logging.InitLogger(assets.NAME_DRSHOW)
	var ErrorOrNil error = nil
	var path string

	mode, err := parseArgs(&path)

	if err != nil {
		FATAL.Exitf("invalid Arguments: %v Resulted in %s", os.Args, err)
		return
	}

	switch mode {
	case mode_pipe:
		INFO.Println("Running in Mode: pipe")
		_, msg := pipe.Run()
		fmt.Println(msg)
	case mode_static:
		if isDirectory(path) {
			INFO.Println("Running in Mode: folder")
			ErrorOrNil = folder.Run(path)
		} else {
			INFO.Println("Running in Mode: file")
			ErrorOrNil = file.Run(path)
		}
		if ErrorOrNil != nil {
			FATAL.Exit(ErrorOrNil)
		}
	default:
		printHelp(path)
	}

}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}
