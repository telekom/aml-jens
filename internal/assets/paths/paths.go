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

var log_path = "/etc/jens-cli/logs/"

// "/etc/jens-cli/logs/"
//
//go:inline
func LOG_PATH() string {
	return log_path
}

//go:inline
func LOG_PATH_UPDATE(p string) {
	log_path = p
}

// "/etc/jens-cli/"
//
//go:inline
func RELEASE_CFG_PATH() string {
	return "/etc/jens-cli/"
}

var testdata = ""

// Only to be used in tests
//
// "{}/test/testdata" (fallback to ./testdata)
//
//go:inline
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

// Only to be used in tests
//
// "{}/test/testdata/drp" (fallback to ./testdata/drp)
//
//go:inline
func TESTDATA_DRP() string {
	t := TESTDATA()
	print(t)
	return filepath.Join(t, "drp/")
}
