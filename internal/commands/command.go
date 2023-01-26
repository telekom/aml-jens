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

package commands

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/telekom/aml-jens/internal/logging"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

// ExecReturnOutput executes 'name' with args
// returns its output in string-form.
// If the program 'name' returns != 0
//
//	!!! --  The Program will CRASH -- !!!
func ExecCrashOnError(name string, arg ...string) {
	err := execCmd(exec.Command(name, arg...))
	go func() {
		if err != nil {
			INFO.Printf("Executing: %s, %v", name, arg)
			FATAL.Printf("Error executing %s", name)
			FATAL.Exit(err)
		}
	}()

}

func execCmdOutput(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		message := fmt.Sprintf("error exec command %s error: %s\n", cmd, err.Error())
		INFO.Print(message)
		INFO.Println(out.String())
		return "", err
	}
	return out.String(), nil
}

func execCmd(cmd *exec.Cmd) error {
	cmd.Stdout = nil

	return cmd.Run()
}

// ExecReturnOutput executes 'name' with args
// returns its output in string-form.
// If the program 'name' returns != 0
// error gets set
func ExecReturnOutput(name string, arg ...string) (string, error) {
	return execCmdOutput(exec.Command(name, arg...))
}
