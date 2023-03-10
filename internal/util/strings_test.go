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

func TestFormatLabelISOKilo(t *testing.T) {
	tests := map[float64]string{
		1.00001:                "1.00k",
		10:                     "10k",
		100000:                 "100M",
		2000000000000000000001: "2Y"}
	for i, expected := range tests {
		got := util.FormatLabelISOKilo(i)
		if got != expected {
			t.Fatalf("Formatting %f; %s!=%s(Got != Expected)", i, got, expected)
		}
	}
}
func TestRemoveWhiteSpace(t *testing.T) {
	if util.RemoveWhiteSpace("  A  B ") != "AB" {
		t.Fatalf("util.RemoveWhiteSpace('  A  B ') -> %s", util.RemoveWhiteSpace("  A  B "))
	}
}
