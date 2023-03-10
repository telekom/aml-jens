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
	"runtime"
	"strings"
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

// Print calls l.Output to print to the logger. Arguments are handled in the manner of fmt.Print.
func (clog *customLogger) Print(v ...any) {
	clog.logger.Print(v...)
	fmt.Fprintln(os.Stderr, "Something went wrong:")
	fmt.Fprint(os.Stderr, v...)
}

// Println calls l.Output to print to the logger. Arguments are handled in the manner of fmt.Println.
func (clog *customLogger) Println(v ...any) {
	clog.logger.Println(v...)
	fmt.Fprintln(os.Stderr, "Something went wrong:")
	fmt.Fprintln(os.Stderr, v...)
}

// Printf calls l.Output to print to the logger. Arguments are handled in the manner of fmt.Printf.
func (clog *customLogger) Printf(format string, v ...any) {
	clog.logger.Printf(format, v...)
	fmt.Fprintln(os.Stderr, "Something went wrong:")
	fmt.Fprintf(os.Stderr, format, v...)
}

// SetOutput sets the output destination for the logger.
func (clog *customLogger) SetOutput(w io.Writer) {
	clog.logger.SetOutput(w)
}

// SetPrefix sets the output prefix for the logger.
func (clog *customLogger) SetPrefix(prefix string) {
	clog.logger.SetPrefix(prefix)
}

// SetFlags sets the output flags for the logger. The flag bits are Ldate, Ltime, and so on.
func (clog *customLogger) SetFlags(flag int) {
	clog.logger.SetFlags(flag)
}

// Writer returns the output destination for the logger.
func (clog *customLogger) Writer() io.Writer {
	return clog.logger.Writer()
}

// Exit calls l.Output to print to the logger and potentially exit. Arguments are handled in the manner of fmt.Print.
func (clog *customLogger) Exit(v ...any) {
	_, file, lineNo, ok := runtime.Caller(1)
	var b strings.Builder
	t := time.Now()
	fmt.Fprintf(&b, "%04d/%02d/%02d %02d:%02d:%02d.%d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	fmt.Fprintf(&b, " %s:%d: [FATAL] %s", file, lineNo, fmt.Sprint(v...))
	for i := 2; i < 10 && ok; i++ {
		_, file, lineNo, ok = runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(&b, "\n    -> %s:%d", file, lineNo)
	}
	b.WriteRune('\n')
	//clog.logger.Print(v...)
	clog.Print(b.String())
	clog.Println("The above exception has caused the program to exit.")
	clog.do_exit()
	clog.Println("Term")
}

// Exitln calls l.Output to print to the logger and potentially exit. Arguments are handled in the manner of fmt.Println.
func (clog *customLogger) Exitln(v ...any) {
	_, file, lineNo, ok := runtime.Caller(1)
	var b strings.Builder
	t := time.Now()
	fmt.Fprintf(&b, "%04d/%02d/%02d %02d:%02d:%02d.%d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	fmt.Fprintf(&b, " %s:%d: [FATAL] %s", file, lineNo, fmt.Sprintln(v...))
	for i := 2; i < 10 && ok; i++ {
		_, file, lineNo, ok = runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(&b, "\n    -> %s:%d", file, lineNo)
	}
	b.WriteRune('\n')
	//clog.logger.Print(v...)
	fmt.Fprintln(
		clog.logger.Writer(),
		b.String())
	clog.logger.Println("The above exception has caused the program to exit.")
	clog.do_exit()
	clog.logger.Println("Term")
}

// Exitf calls l.Output to print to the logger and potentially exit. Arguments are handled in the manner of fmt.Printf.
func (clog *customLogger) Exitf(format string, v ...any) {
	_, file, lineNo, ok := runtime.Caller(1)
	var b strings.Builder
	b.Grow(60)
	t := time.Now()
	fmt.Fprintf(&b, "%04d/%02d/%02d %02d:%02d:%02d.%d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	fmt.Fprintf(&b, " %s:%d: [FATAL] %s", file, lineNo, fmt.Sprintf(format, v...))
	for i := 2; i < 10 && ok; i++ {
		_, file, lineNo, ok = runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(&b, "\n    -> %s:%d", file, lineNo)
	}
	b.WriteRune('\n')
	//clog.logger.Print(v...)
	fmt.Fprintln(
		clog.logger.Writer(),
		b.String())
	clog.logger.Println("The above exception has caused the program to exit.")
	clog.do_exit()
	clog.logger.Println("Term")
}

// Calls the set callback (exit); and waits for a set amount of time
// After the wait time os.Exit(-1) is called.
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
	res.setExitFunc(noop, 0x2)
	return res
}
