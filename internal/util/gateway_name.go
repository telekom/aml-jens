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
package util

import (
	"fmt"
	"net"
	"strings"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/commands"
)

func get_user() string {
	res := commands.ExecCommand("bash", "-c", assets.CMD_GET_USERS)

	if res.Error() == nil {
		lines := strings.Split(res.StdOut(), "\n")
		lines = lines[:len(lines)-1]
		switch cnt := len(lines); {
		case cnt == 1:
			l := strings.SplitAfter(lines[0], " ")

			return l[0]
		case cnt > 1:
			return lines[1]
		default:
			return "jens"
		}
	}
	return "jens"
}

func get_gateway(usr string) string {
	if !isIpOrJens(usr) {
		INFO.Printf("Invalid usr: %s, defaulting\n", usr)
		return "jens"
	}
	res := commands.ExecCommand("bash", "-c", fmt.Sprintf(assets.CMD_GET_GATEWAY, usr))
	if err := res.Error(); err != nil {
		INFO.Printf("Could not determine Gateway: %s\n", err)
		return "localhost"
	}
	return strings.TrimRight(res.StdOut(), "\n")
}

func isIpOrJens(adr string) bool {
	return net.ParseIP(adr) != nil || adr == "jens"
}

// Returns educated guess, on which ip the grafana is reachable
func RetrieveMostLikelyGatewayIp() string {
	usr := get_user()
	if usr == "-" {
		usr = "localhost"
	}
	gw := get_gateway(usr)
	if isIpOrJens(gw) {
		INFO.Printf("'%s' is not a valid gateway_ip -> defaulting\n", gw)
		gw = "jens"
	}
	return gw
}
