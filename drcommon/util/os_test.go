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

package util_test

import (
	"jens/drcommon/util"
	"os"
	"testing"
)

func TestIsDirectory(t *testing.T) {
	isdir, err := util.IsDirectory("/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !isdir {
		t.Fatal("/tmp was not interpreted as a directory")
	}
}

func TestIsDirectoryFalse(t *testing.T) {
	path := "/tmp/file"
	os.Create(path)
	isdir, err := util.IsDirectory(path)
	if err != nil {
		t.Fatal(err)
	}
	if isdir {
		t.Fatal("/tmp/file was interpreted as a directory")
	}
	os.Remove(path)
}
