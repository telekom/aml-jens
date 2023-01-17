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

package acquisition

import (
	"fmt"
	"path/filepath"

	dm "github.com/telekom/aml-jens/cmd/drshow/internal/data/drpdata"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

func NewDrpListFromFolder(path string) (*dm.DrpListT, error) {
	folders, err := filepath.Glob(fmt.Sprintf("%s/*.csv", path))
	if err != nil {
		return nil, err
	}
	res := dm.NewDrpListT()
	for i := 0; i < len(folders); i++ {
		drp, err := NewDrpFromFile(folders[i])
		if err != nil {
			continue
		}
		res.Drps = append(res.Drps, *drp)
	}
	if len(res.Drps) > 0 {
		res.Selected = 0
	}
	return res, nil
}
func NewDrpFromFile(path string) (*dm.DrpT, error) {
	drp := datatypes.DB_data_rate_pattern{
		Scale:        1,
		MinRateKbits: 1,
	}

	if err := drp.ParseDrpFile(path); err != nil {
		return nil, err
	}
	min, max, avg := drp.GetStats()
	drp2 := dm.NewDrpT(path, min, max, avg)
	data := drp.GetInternalPattern()
	drp2.SetData(data)
	return drp2, nil

}
