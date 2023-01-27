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
	"fmt"
	"path/filepath"

	"github.com/telekom/aml-jens/internal/errortypes"
)

// Will skip faulty DRPs
// Does not apply any mirate restricitons
func NewDrpListFromFolder(path string) (*DrpListT, error) {
	folders, err := filepath.Glob(fmt.Sprintf("%s/*.csv", path))
	if err != nil {
		return nil, err
	}
	providers := make([]DataRatePatternProvider, len(folders))
	for i, path := range folders {
		providers[i] = NewDataRatePatternFileProvider(path)

	}
	return NewDrpList(providers,
		[]struct {
			scale     float64
			minrateKb float64
		}{
			{
				scale:     1,
				minrateKb: 0,
			},
		},
		true)
}
func NewDrpList(providers []DataRatePatternProvider, args []struct {
	scale     float64
	minrateKb float64
}, skipOnFaultyDrp bool) (*DrpListT, error) {
	if len(providers) != len(args) && len(args) != 1 {
		return nil, errortypes.NewUserInputError("Length of providers and args do not match.")
	}
	res := NewDrpListT()
	for i, v := range providers {
		var drp DataRatePattern
		var err error
		if len(args) == 1 {
			drp, err = v.Provide(args[0].scale, args[0].minrateKb)
		} else {
			drp, err = v.Provide(args[i].scale, args[i].minrateKb)
		}
		if err != nil {
			if !skipOnFaultyDrp {
				return nil, errortypes.NewUserInputError("Got a faluty DRP: %+v", err)
			} else {
				continue
			}
		}
		res.Drps = append(res.Drps, drp)
	}
	res.Selected = 0
	if len(res.Drps) <= 0 {
		return nil, errortypes.NewUserInputError("No DRPs were loaded")
	}
	return res, nil
}
