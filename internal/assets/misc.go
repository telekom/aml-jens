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

package assets

import "log"

const (
	NAME_DRPLAY   = "DrPlay"
	NAME_DRSHOW   = "DrShow"
	NAME_DRBENCH  = "DrBenchmark"
	LOG_PRE_DEBUG = "[DEBUG] "
	LOG_PRE_INFO  = "[INFO] "
	LOG_PRE_WARN  = "[WARN] "
	LOG_PRE_FATAL = "[FATAL] "
	LOG_SETTING   = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.Lmsgprefix
)

const (
	CMD_GET_USERS   = "w --no-header --short | awk '{print $3}'"
	CMD_GET_GATEWAY = "ip --brief route get %s | rev | awk '{print $3}' | rev"
)

// Heading for stdout of drplay --> stdin for drshow.pipe
//
// [timestamp, soj, load, ...]
var CONST_HEADING = []string{"timestampMs", "sojournTimeMs", "loadKbits", "capacityKbits", "ecnCePercent", "dropped", "netflow"}

var END_OF_DRPLAY = [...]string{"data", "rate", "player", "ended"}

const (
	NFT_TABLE_PREMARK = "premarkect1"
	NFT_TABLE_SIGNAL  = "signalect0"
	NFT_CHAIN_FORWARD = "forward"
	NFT_CHAIN_OUTPUT  = "output"
)

/*
	The following are set using ldbuild flags
*/
var VERSION string = "NaN"
var BUILD_TIME string = "0"
