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
	"encoding/hex"
	"errors"
	"fmt"
	"jens/drcommon/persistence/datatypes"
	"jens/drcommon/util"
	"os"
)

type BenchmarkPattern struct {
	Path    string
	hash    []byte
	HashStr string `json:"Hash"`
	Setting DrplaySetting
	pattern *datatypes.DB_data_rate_pattern
}

func NewBenchmarkPattern(path string, setting DrplaySetting) BenchmarkPattern {
	b := BenchmarkPattern{
		Path:    path,
		Setting: setting,
	}
	b.checkOrCalcHash()
	return b
}

func (bp *BenchmarkPattern) GetDrp() *datatypes.DB_data_rate_pattern {
	return bp.pattern
}
func (bp *BenchmarkPattern) loadPattern() error {
	if bp.Path == "" {
		return errors.New("can't load a pattern without a path")
	}
	bp.pattern = &datatypes.DB_data_rate_pattern{
		MinRateKbits: 0.9652,
		Scale:        0.9999965,
		Loop:         false,
	}
	return bp.pattern.ParseDrpFile(bp.Path)
}

func (bp *BenchmarkPattern) GetHashOfLoadedPattern() []byte {
	h := bp.pattern.GetHash()
	return h
}
func (bp *BenchmarkPattern) checkOrCalcHash() error {
	if bp.pattern == nil {
		//forward err
		if err := bp.loadPattern(); err != nil {
			return err
		}
	}
	if len(bp.hash) == 0 {
		INFO.Println("Hash of BMP not set while Validating. Generating")
		bp.hash = bp.GetHashOfLoadedPattern()
	}
	calced_hash := hex.EncodeToString(bp.hash)
	if bp.HashStr != "" {
		if calced_hash != bp.HashStr {
			return fmt.Errorf("DRP-Hash is incorrect. Cant Use the pattern %s (%s) as (%s)", bp.Path, calced_hash, bp.HashStr)
		}
		return nil
	}
	bp.HashStr = calced_hash

	return nil
}
func (bp *BenchmarkPattern) Validate() error {
	err := bp.checkOrCalcHash()
	if err != nil {
		return err
	}
	E := func(s string) error { return fmt.Errorf("BenchmarkPattern: %s", s) }
	if _, err := os.Stat(bp.Path); errors.Is(err, os.ErrNotExist) {
		return E(fmt.Sprintf("File %s does not exit", bp.Path))
	}
	//TODO: add argument for minratekbits

	if len(bp.hash) == 0 {
		INFO.Println("Hash of BMP not set while Validating. Generating")
		bp.hash = bp.GetHashOfLoadedPattern()
	} else {
		loadedHash := bp.GetHashOfLoadedPattern()
		if !util.ByteCompare(bp.hash, loadedHash) {
			return E(fmt.Sprintf("DRP-Hash is incorrect. Cant Use the pattern %s (%s) as (%s)",
				bp.Path, bp.pattern.GetHashStr(), bp.HashStr))
		}
	}
	return bp.Setting.Validate()
}
