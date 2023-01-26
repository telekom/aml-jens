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

package logging

import (
	"fmt"
	"io"
	rlog "log"
	"os"
	"path/filepath"
	"strings"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/assets/paths"
)

type LoggerType interface {
	Prefix() string

	Print(v ...any)
	Println(v ...any)
	Printf(format string, v ...any)
	SetOutput(w io.Writer)
	SetPrefix(prefix string)
	SetFlags(flag int)
	Writer() io.Writer
	Exit(v ...any)
	Exitln(v ...any)
	Exitf(format string, v ...any)
	setExitFunc(new_exit func() uint8, timeoutMs int)
}

var (
	singelton_debug_logger *rlog.Logger = nil
	singelton_info_logger  *rlog.Logger = nil
	singelton_warn_logger  *rlog.Logger = nil
	singelton_fatal_logger LoggerType   = nil
	program_name           string       = "dr"
	fp                     *os.File     = nil
)

type dualWriter struct {
	file *os.File
	out  *os.File
}

func NewDualWriter(file *os.File, out *os.File) *dualWriter {
	return &dualWriter{
		file: file,
		out:  out,
	}
}
func (s *dualWriter) Write(p []byte) (n int, err error) {
	n, err = s.file.Write(p)
	if err != nil {
		return n, err
	}
	start := 0
	for i, v := range p {
		if v == byte(']') {
			start = (i + 1)
			break
		}
	}
	n, err = s.out.Write([]byte("An error occurred:"))
	if err != nil {
		return n, err
	}
	n, err = s.out.Write(p[start:])
	return n, err
}

func InitLogger(name string) {
	var err error
	err = os.MkdirAll(paths.LOG_PATH(), 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "COULD NOT INIT LOGGER")
		fmt.Fprint(os.Stderr, err)
		os.Exit(-1)
	}
	if name != "" {
		program_name = strings.ReplaceAll(name, "/", "_")
	}
	file := program_name + ".log"

	path := filepath.Join(paths.LOG_PATH(), file)
	fp, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while creating Logger")
		fmt.Fprint(os.Stderr, err)
		os.Exit(-1)
	}
	singelton_info_logger.SetOutput(fp)
	singelton_warn_logger.SetOutput(fp)
	singelton_fatal_logger.SetOutput(NewDualWriter(fp, os.Stderr))
}
func EnableDebug() {
	singelton_debug_logger.SetOutput(fp)
}
func LinkExitFunction(exit func() uint8, timeoutMs int) {
	if singelton_fatal_logger != nil {
		singelton_fatal_logger.setExitFunc(exit, timeoutMs)
	} else {
		rlog.Default().Fatal("Cant set Exit funtion on not set Logger")
	}
}
func GetLogger() (debug *rlog.Logger, info *rlog.Logger, warn *rlog.Logger, fatal LoggerType) {
	if singelton_fatal_logger == nil || singelton_info_logger == nil {
		singelton_debug_logger = (rlog.New(io.Discard, assets.LOG_PRE_DEBUG, assets.LOG_SETTING))
		singelton_info_logger = (rlog.New(os.Stderr, assets.LOG_PRE_INFO, assets.LOG_SETTING))
		singelton_fatal_logger = NewCustomLogger(rlog.New(os.Stderr, assets.LOG_PRE_FATAL, assets.LOG_SETTING))
		singelton_warn_logger = (rlog.New(os.Stderr, assets.LOG_PRE_WARN, assets.LOG_SETTING))
	}
	return singelton_debug_logger, singelton_info_logger, singelton_warn_logger, singelton_fatal_logger
}
