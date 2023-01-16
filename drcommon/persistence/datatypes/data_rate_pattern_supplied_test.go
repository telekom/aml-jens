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

package datatypes_test

import (
	"io/ioutil"
	"jens/drcommon/assets"
	"jens/drcommon/persistence/datatypes"
	"strings"
	"testing"
)

func getDrps(t *testing.T, path string) []string {
	res := make([]string, 0, 6)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		name := v.Name()
		if strings.HasPrefix(name, "drp_") &&
			strings.HasSuffix(name, ".csv") {
			if strings.HasSuffix(name, "comment.csv") {
				continue
			}
			res = append(res, name)
		}

	}
	return res
}

func TestStdDrp(t *testing.T) {
	var hashes = map[string]string{
		"drp_3valleys.csv":        "020b6fc00d7a6f91c050a0833f11c18d",
		"drp_bike.csv":            "ac820f3c891f46021cb40598305897d0",
		"drp_ericsson.csv":        "090f44a522a84f6dc2b1655c4ac17b76",
		"drp_munich_autobahn.csv": "0d271f1acdc73768f9df24eeb7f633dc",
		"drp_munich_outskirt.csv": "832dcb6836a5eec870d2db66b5c77956",
		"drp_munich_village.csv":  "394fa74338d447539e9d0b7e52240a99",
	}
	drps := getDrps(t, assets.TESTDATA_DPR_DIR)
	if len(drps) == 0 {
		t.Fatal("No std DRPs found.")
	}
	for _, v := range drps {
		drp := datatypes.DB_data_rate_pattern{
			Scale: 1,
			Loop:  false,
		}
		err := drp.ParseDrpFile(assets.TESTDATA_DPR_DIR + v)
		if err != nil {
			t.Fatal(err)
		}
		if err != nil {
			t.Fatalf("Loaded std drp, got an error: %s", err)
		}
		if drp.GetHashStr() != hashes[v] {
			t.Logf("Got: %s != %s", drp.GetHashStr(), hashes[v])
			t.Fatalf("Pattern %s seems to have changed", v)
		}
	}
}
