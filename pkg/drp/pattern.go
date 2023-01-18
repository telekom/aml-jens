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

package drp

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"strings"
)

type DataRatePattern struct {
	iter            *DataRatePatternIterator
	Name            string
	data            *[]float64
	Filepath        string
	Description     string
	Mapping         map[string]string
	Min             float64
	Max             float64
	Avg             float64
	Length          int
	Sha256          []byte
	Th_mq_latency   string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_p95_latency  string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_p99_latency  string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_p999_latency string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_link_usage   string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
}

func NewDataRatePattern(filepath string, min float64, max float64, avg float64) *DataRatePattern {
	filenameS := strings.Split(filepath, "/")
	filename := strings.Split(filenameS[len(filenameS)-1], ".")

	return &DataRatePattern{
		Name:     strings.Join(filename[:len(filename)-1], ""),
		Filepath: filepath,
		Min:      min,
		Max:      max,
		Avg:      avg,
		iter:     nil,
	}
}
func (s *DataRatePattern) SampleCount() int {
	return len(*s.data)
}
func (s *DataRatePattern) SetData(d []float64) {
	cpy := make([]float64, len(d))
	copy(cpy, d)
	s.data = &cpy
}
func (s *DataRatePattern) GetData() *[]float64 {
	return s.data
}
func (s *DataRatePattern) GetHash() []byte {
	return s.Sha256
}
func (s *DataRatePattern) GetHashStr() string {
	return hex.EncodeToString(s.Sha256)
}
func (s *DataRatePattern) GetStats() (min float64, max float64, avg float64) {
	return s.Min, s.Max, s.Avg
}

func (s *DataRatePattern) Iterator() *DataRatePatternIterator {
	if s.iter == nil {
		s.iter = NewDataRatePatternIterator(s)
	}
	return s.iter
}

// Resacle all values. This also re-calculates the hash
func (s *DataRatePattern) ScaleByInt(scale int) error {
	return s.Scale(float64(scale))
}

// Resacle all values. This also re-calculates the hash
func (s *DataRatePattern) Scale(scale float64) error {
	var hash_buf bytes.Buffer
	for i, v := range *s.data {
		newv := scale * v
		(*s.data)[i] = newv
		binary.Write(&hash_buf, binary.LittleEndian, newv)
	}
	hash := md5.New()
	if _, err := hash.Write(hash_buf.Bytes()); err != nil {
		return err
	}
	s.Sha256 = hash.Sum(nil)
	s.Avg *= scale
	s.Max *= scale
	s.Min *= scale
	return nil
}
