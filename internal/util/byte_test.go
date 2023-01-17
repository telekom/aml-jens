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
	"testing"

	"github.com/telekom/aml-jens/internal/util"
)

func TestByteCompareTrue(t *testing.T) {
	if util.ByteCompare([]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6}, []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6}) == false {
		t.Fatal("Two equal byte-Arrays compared untrue.")
	}
}

func TestByteCompareFals(t *testing.T) {
	if util.ByteCompare([]byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0xf}, []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7}) == true {
		t.Fatal("Two not equal byte-Arrays compared tue.")
	}
}
