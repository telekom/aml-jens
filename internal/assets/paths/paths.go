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

package paths

import (
	"path/filepath"
	"runtime"
)

func LOG_PATH() string {
	return "/etc/jens-cli/logs/"
}
func RELEASE_CFG_PATH() string {
	return "/etc/jens-cli/"
}

var testdata = ""

func TESTDATA() string {
	if testdata == "" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			return "./testdata"
		}
		testdata = filepath.Join(filepath.Dir(filename), "../../../test/testdata")
	}

	return testdata
}
func TESTDATA_DRP() string {
	t := TESTDATA()
	return filepath.Join(t, "drp/")
}
