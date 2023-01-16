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

package trafficcontrol

import (
	"bytes"
	"jens/drcommon/commands"
	"os/exec"
)

func CreateNftRuleECT(dev string, nftTable string, chainForward string, chainOutput string, ect string, priority string) {
	commands.ExecCrashOnError("nft", "add", "table", "inet", nftTable)

	commands.ExecCrashOnError("nft", "add", "chain", "inet", nftTable, chainForward, "{", "type", "filter", "hook", "forward", "priority", priority, ";}")

	commands.ExecCrashOnError("nft", "add", "chain", "inet", nftTable, chainOutput, "{", "type", "filter", "hook", "output", "priority", priority, ";}")

	commands.ExecCrashOnError("nft", "add", "rule", "inet", nftTable, chainForward, "ip", "ecn", "set", ect, "oifname", dev)
	commands.ExecCrashOnError("nft", "add", "rule", "inet", nftTable, chainOutput, "ip", "ecn", "set", ect, "oifname", dev)

	INFO.Printf("enabled nft rules for %s %s", nftTable, ect)
}

func ResetECTMarking(nftTable string) {
	cmd := exec.Command("nft", "delete", "table", "inet", nftTable)
	runCommand(cmd)
	INFO.Printf("reset nft ect1 %s", nftTable)
}

func runCommand(cmd *exec.Cmd) {
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		INFO.Printf("%s -> %s; NOT CRASHING.", cmd.Args, err)
	}
}
