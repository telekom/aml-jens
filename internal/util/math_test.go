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
	"log"
	"testing"

	"github.com/telekom/aml-jens/internal/util"
)

func TestFloorToInt(t *testing.T) {
	if util.FloorToInt(1.9) != 1 {
		log.Fatalf("1.9 was not floored to 1. (was: %d)", util.FloorToInt(1.9))
	}
}

func TestMaxInt(t *testing.T) {
	if util.MaxInt(99, 991) != 991 {
		t.Fatal("991 was not chosen as Max(99,991)")
	}
}
func TestMinInt(t *testing.T) {
	if util.MinInt(99, 991) != 99 {
		t.Fatal("99 was not chosen as Min(99,991)")
	}
}
