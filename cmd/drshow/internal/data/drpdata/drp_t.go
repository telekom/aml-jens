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

package drpdata

import "strings"

type DrpT struct {
	Name     string
	data     *[]float64
	Filepath string
	Min      float64
	Max      float64
	Avg      float64
}

func NewDrpT(filepath string, min float64, max float64, avg float64) *DrpT {
	filenameS := strings.Split(filepath, "/")
	filename := strings.Split(filenameS[len(filenameS)-1], ".")

	return &DrpT{
		Name:     strings.Join(filename[:len(filename)-1], ""),
		Filepath: filepath,
		Min:      min,
		Max:      max,
		Avg:      avg,
	}
}
func (s *DrpT) SampleCount() int {
	return len(*s.data)
}
func (s *DrpT) SetData(d []float64) {
	cpy := make([]float64, len(d))
	copy(cpy, d)
	s.data = &cpy
}
func (s *DrpT) GetData() *[]float64 {
	return s.data
}
