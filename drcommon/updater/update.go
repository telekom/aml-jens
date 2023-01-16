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

package updater

import (
	"fmt"
	"log"
)

func DisplayUpdateMsgIfNewerVersion() {
	if !updateAvailableOnApt() {
		return
	}
	log.Println("An update is available. Check Apt-get")
	fmt.Print(
		"-------------------------------------------\n" +
			"| An update for jens-cli is available.    |\n" +
			"| Run 'apt list --upgradable jens-cli' to |\n" +
			"| review the update.                      |\n" +
			"| Alternatively you can update the entire |\n" +
			"| JENS using apt.                         |\n" +
			"-------------------------------------------\n",
	)
}
