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
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func FormatLabelISOKilo(nbr float64) string {
	return FormatLabelISO(nbr * 1000)
}
func formatLabelISO(nbr float64, formatter func(float64, string) string) string {
	if nbr > 1000 {
		if nbr >= 1000000 {
			if nbr >= 1000000000 {
				if nbr >= 1000000000000 {
					if nbr >= 1000000000000000 {
						if nbr >= 1000000000000000000 {
							if nbr >= 1000000000000000000000 {
								if nbr >= 1000000000000000000000000 {

									return formatter(nbr/1000000000000000000000000, "Y")
								}

								return formatter(nbr/1000000000000000000000, "Z")
							}

							return formatter(nbr/1000000000000000000, "E")
						}

						return formatter(nbr/1000000000000000, "P")
					}

					return formatter(nbr/1000000000000, "T")
				}

				return formatter(nbr/1000000000, "G")
			}

			return formatter(nbr/1000000, "M")
		}

		return formatter(nbr/1000, "k")
	}
	return formatter(nbr, "")
}
func FormatLabelISO(nbr float64) string {
	noTrailingZeros := func(i float64, s string) string {
		if i == float64(int64(i)) { //no decimals
			return fmt.Sprintf("%.0f%  s", i, s)
		} else { //decimals
			return fmt.Sprintf("%.2f%s", i, s)
		}
	}
	return formatLabelISO(nbr, noTrailingZeros)
}
func FormatLabelISOShorter(nbr float64) string {
	noTrailingZeros := func(i float64, s string) string {
		if i == float64(int64(i)) { //no decimals
			return fmt.Sprintf("%.0f%  s", i, s)
		} else { //decimals
			return fmt.Sprintf("%.1f%s", i, s)
		}
	}
	return formatLabelISO(nbr, noTrailingZeros)
}

func RemoveWhiteSpace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func IterateTagName(name string) string {
	match, _ := regexp.MatchString(".*\\([0-9]+\\)", name)
	if match {
		split := strings.Split(name, "(")
		si := split[len(split)-1]
		i, err := strconv.Atoi(si[:len(si)-1])
		if err != nil {
			WARN.Println("Could not mangle name for unique entry")
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(time.Now().Unix()))
			return name + base64.StdEncoding.EncodeToString(b)
		}
		name = fmt.Sprintf("%s(%d)", split[0], i+1)
	} else {
		name += "(1)"
	}
	return name
}
