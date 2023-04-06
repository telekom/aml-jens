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

type CommandResult struct {
	s_out       string
	s_err       string
	err         error
	command_str string
}

//go:inline
func (s CommandResult) StdOut() string {
	return s.s_out
}

//Will return an error if result was not successful
//Else: nil
//go:inline
func (s CommandResult) Error() error {
	if s.err == nil {
		return nil
	}
	return fmt.Errorf("command '%s', [stdout:'%s', stderr:'%s'] failed: %w", s.command_str, s.s_out, s.s_err, s.err)
}

// ExecReturnOutput executes 'name' with args
// returns CommandResult.
func ExecCommand(name string, arg ...string) CommandResult {
	return ExecCommandEnv(name, []string{}, arg...)
}

// ExecReturnOutput executes 'name' with args
// returns CommandResult.
func ExecCommandEnv(name string, env []string, arg ...string) CommandResult {
	cmd := exec.Command(name, arg...)
	cmd.Env = append(cmd.Env, env...)
	return execCmd(cmd)

}

func execCmd(cmd *exec.Cmd) CommandResult {
	var out bytes.Buffer
	var outE bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outE
	err := cmd.Run()
	return CommandResult{
		err:         err,
		s_out:       out.String(),
		s_err:       outE.String(),
		command_str: cmd.String(),
	}
}
