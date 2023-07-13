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
	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/commands"
	"strconv"
)

const CHAIN_UEMARK_FORWARD = "uemarkforward"
const CHAIN_UEMARK_OUTPUT = "uemarkoutput"

func CreateRuleECT(dev string, nftTable string, chainForward string, chainOutput string, ect string, priority string) error {
	if res := commands.ExecCommand("nft", "add", "table", "inet", nftTable); res.Error() != nil {
		return res.Error()
	}

	if res := commands.ExecCommand("nft", "add", "chain", "inet", nftTable, chainForward, "{", "type", "filter", "hook", "forward", "priority", priority, ";}"); res.Error() != nil {
		return res.Error()
	}

	if res := commands.ExecCommand("nft", "add", "chain", "inet", nftTable, chainOutput, "{", "type", "filter", "hook", "output", "priority", priority, ";}"); res.Error() != nil {
		return res.Error()
	}

	if res := commands.ExecCommand("nft", "add", "rule", "inet", nftTable, chainForward, "ip", "ecn", "set", ect, "oifname", dev); res.Error() != nil {
		return res.Error()
	}

	if res := commands.ExecCommand("nft", "add", "rule", "inet", nftTable, chainOutput, "ip", "ecn", "set", ect, "oifname", dev); res.Error() != nil {
		return res.Error()
	}

	DEBUG.Printf("enabled nft rules for %s %s", nftTable, ect)
	return nil
}

func ResetNFT(nftTable string) {
	res := commands.ExecCommand("nft", "delete", "table", "inet", nftTable)
	err := res.Error()
	if err != nil {
		WARN.Println(err)
	}
}

func CreateRulesMarkUe(Netflows []string) error {
	if res := commands.ExecCommand("nft", "add", "table", "inet", assets.NFT_TABLE_UEMARK); res.Error() != nil {
		return res.Error()
	}

	if res := commands.ExecCommand("nft", "add", "chain", "inet", assets.NFT_TABLE_UEMARK, CHAIN_UEMARK_FORWARD, "{", "type", "filter", "hook", "forward", "priority", "0", ";}"); res.Error() != nil {
		return res.Error()
	}

	if res := commands.ExecCommand("nft", "add", "chain", "inet", assets.NFT_TABLE_UEMARK, CHAIN_UEMARK_OUTPUT, "{", "type", "filter", "hook", "output", "priority", "0", ";}"); res.Error() != nil {
		return res.Error()
	}
	for i := 0; i < len(Netflows); i++ {
		netflow := Netflows[i]
		//netflowArray := strings.Split(netflow, " ")

		if res := commands.ExecCommand("nft", "add", "rule", "inet", assets.NFT_TABLE_UEMARK, CHAIN_UEMARK_FORWARD, netflow, "meta", "mark", "set", strconv.Itoa(i+1), "counter"); res.Error() != nil {
			return res.Error()
		}
		if res := commands.ExecCommand("nft", "add", "rule", "inet", assets.NFT_TABLE_UEMARK, CHAIN_UEMARK_OUTPUT, netflow, "meta", "mark", "set", strconv.Itoa(i+1), "counter"); res.Error() != nil {
			return res.Error()
		}
	}

	DEBUG.Printf("enabled nft rules for %s", assets.NFT_TABLE_UEMARK)
	return nil
}
