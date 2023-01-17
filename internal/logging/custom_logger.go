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
	"io"
	rlog "log"
	"os"
	"time"
)

type customLogger struct {
	logger  *rlog.Logger
	exit    func() uint8
	timeout time.Duration
}

func (clog *customLogger) setExitFunc(new_exit func() uint8, timeoutMs int) {
	clog.exit = new_exit
	clog.timeout = time.Millisecond * time.Duration(timeoutMs)
}

func (clog *customLogger) Prefix() string {
	return clog.logger.Prefix()
}
func (clog *customLogger) Print(v ...any) {
	clog.logger.Print(v...)
}
func (clog *customLogger) Println(v ...any) {
	clog.logger.Println(v...)
}
func (clog *customLogger) Printf(format string, v ...any) {
	clog.logger.Printf(format, v...)
}
func (clog *customLogger) SetOutput(w io.Writer) {
	clog.logger.SetOutput(w)
}
func (clog *customLogger) SetPrefix(prefix string) {
	clog.logger.SetPrefix(prefix)
}
func (clog *customLogger) SetFlags(flag int) {
	clog.logger.SetFlags(flag)
}
func (clog *customLogger) Writer() io.Writer {
	return clog.logger.Writer()
}
func (clog *customLogger) Exit(v ...any) {
	clog.logger.Print(v...)
	clog.do_exit()
}
func (clog *customLogger) Exitln(v ...any) {
	clog.logger.Println(v...)
	clog.do_exit()
}
func (clog *customLogger) Exitf(format string, v ...any) {
	clog.logger.Printf(format, v...)
	clog.do_exit()
}

func (clog *customLogger) do_exit() {
	c := make(chan uint8, 1)
	go func() {
		res := clog.exit()
		c <- res
	}()
	select {
	case <-c:
		return
	case <-time.After(clog.timeout):
		clog.Println("Ungraceful shutdown: Timout Exceeded")
		os.Exit(-1)
	}
}

func NewCustomLogger(logger *rlog.Logger) *customLogger {
	noop := func() uint8 { return 3 }
	res := &customLogger{logger: logger}
	res.setExitFunc(noop, 0xffff)
	return res
}
