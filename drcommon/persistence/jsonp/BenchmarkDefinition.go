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

package jsonp

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

type BenchmarkDefinition struct {
	Name                      string
	Max_application_bitrate   int
	MaxBitrateEstimationTimeS int
	Patterns                  []BenchmarkPattern
	DrplaySetting             DrplaySetting
}

func (bcfg *BenchmarkDefinition) CalcMd5() ([]byte, error) {
	h := md5.New()
	data, err := json.Marshal(*bcfg)
	if err != nil {
		return nil, err
	}
	h.Write(data)
	return h.Sum(nil), nil
}

func (bcfg *BenchmarkDefinition) Validate() error {
	E := func(s string) error { return fmt.Errorf("BenchmarkCfg: %s", s) }
	if bcfg.Name == "" {
		return E("name is not set")
	}
	/*if bcfg.Max_application_bitrate < 999 {
		return E(fmt.Sprintf("'Max_application_bitrate' needs to be >= 1000, is %d", bcfg.Max_application_bitrate))
	}*/

	if bcfg.MaxBitrateEstimationTimeS == 0 {
		return E("MaxBitrateEstimationTimeS not set or equal to 0")
	}
	if len(bcfg.Patterns) == 0 {
		return E("No Patterns in loaded file")
	}
	for i := range bcfg.Patterns {
		err := bcfg.Patterns[i].Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
