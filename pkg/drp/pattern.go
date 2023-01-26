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
	iter           *DataRatePatternIterator
	Name           string
	data           *[]float64
	Description    string
	mapping        map[string]string
	Min            float64
	Max            float64
	Avg            float64
	Length         int
	Sha256         []byte
	loadParameters struct {
		MinRateKbits float64
		Scale        float64
		Origin       string
	}
}

// Create a new DataRatePattern object.
// Most important members get initialized.
//
// No files are read. - FileProviders read files
//
// NOTE: this function might become internal in the future
func NewDataRatePattern(params struct {
	MinRateKbits float64
	Scale        float64
	Origin       string
}) *DataRatePattern {
	filenameS := strings.Split(params.Origin, "/")
	filename := strings.Split(filenameS[len(filenameS)-1], ".")

	return &DataRatePattern{
		Name:           strings.Join(filename[:len(filename)-1], ""),
		loadParameters: params,
		Min:            0,
		Max:            0,
		Avg:            0,
		iter:           nil,
		mapping: map[string]string{
			"th_mq_latency":   "{2,4}",
			"th_p95_latency":  "{10,20}",
			"th_p99_latency":  "{10,20}",
			"th_p999_latency": "{10,20}",
			"th_link_usage":   "{60,80}",
		},
	}
}

// Return KeyValue parameter specified by pattern or fallback
func (s *DataRatePattern) GetMappingValue(name string, fallback string) string {
	v, found := s.mapping[name]
	if !found {
		v = fallback
	}
	return v
}

// Returns the amount of samples in the pattern
//
//go:inline
func (s *DataRatePattern) SampleCount() int {
	return len(*s.data)
}

// Returns the set minimum rate.
//
// Note: this is not the minimum of the pattern.
// It might be higher. Use GetStats() instead!
//
//go:inline
func (s *DataRatePattern) GetMinRateKbits() float64 {
	return s.loadParameters.MinRateKbits
}

// Returns the set scale
//
//go:inline
func (s *DataRatePattern) GetScale() float64 {
	return s.loadParameters.Scale
}

// Returns the origin. (Value set by provider)
//
//go:inline
func (s *DataRatePattern) GetOrigin() string {
	return s.loadParameters.Origin
}

// Sets the internal pattern to d
func (s *DataRatePattern) SetData(d []float64) {
	cpy := make([]float64, len(d))
	copy(cpy, d)
	s.data = &cpy
	s.iter.UpdateAndReset(s.data)
}

// Returns the set data
//
// NOTE: changes to this will not change stats.
func (s *DataRatePattern) GetData() *[]float64 {
	return s.data
}

// Returns Sha256 hash as a byte array
//
//go:inline
func (s *DataRatePattern) GetHash() []byte {
	return s.Sha256
}

// Returns Sha256 hash as a hey encoded string
//
//go:inline
func (s *DataRatePattern) GetHashStr() string {
	return hex.EncodeToString(s.Sha256)
}

// Returns Minumum, Maximum, Average of the loaded pattern
//
//go:inline
func (s *DataRatePattern) GetStats() (min float64, max float64, avg float64) {
	return s.Min, s.Max, s.Avg
}

// Creates / Returns a Iterator.
//
// The Iterator is bound to this object.
func (s *DataRatePattern) Iterator() *DataRatePatternIterator {
	if s.iter == nil {
		s.iter = NewDataRatePatternIterator()
		if s.data != nil && len(*s.data) != 0 {
			s.iter.UpdateAndReset(s.data)
		}
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
		err := binary.Write(&hash_buf, binary.LittleEndian, newv)
		if err != nil {
			return err
		}
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
