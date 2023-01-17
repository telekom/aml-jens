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
	"testing"
)

func TestIsJensValidIp(t *testing.T) {
	if !isIpOrJens("jens") {
		t.Fatal("Jens was not recognized as a valid ip")
	}
}

func TestIsInValidIp(t *testing.T) {
	if isIpOrJens("192.168.2.2.2") {
		t.Fatal("Invalid Ip was recognized as correct")
	}
}
func TestIsEmptyStr(t *testing.T) {
	if isIpOrJens("") {
		t.Fatal("Empty string was recognized as correct")
	}
}
func TestIsValidIp(t *testing.T) {
	if !isIpOrJens("192.168.2.2") {
		t.Fatal("Vaid Ip was not recognized as correct")
	}
}
