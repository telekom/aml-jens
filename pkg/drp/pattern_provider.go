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
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/util"
)

var DEBUG, INFO, FATAL = logging.GetLogger()

type DataRatePatternProvider interface {
	Provide() DataRatePattern
}

type DataRatePatternFileProvider struct {
	Path string
}

func NewDataRatePatternFileProvider(path string) *DataRatePatternFileProvider {
	return &DataRatePatternFileProvider{Path: path}
}
func convertDRPdata(strdata *[][]string) (drp DataRatePattern, err error) {
	if len(*strdata) == 0 {
		return drp,
			errortypes.NewUserInputError("DRP seems to be invalid. No rows loaded.")
	}
	drp.Min = math.MaxFloat64
	drp.Max = -1
	drp.Avg = 0
	drp.Length = len(*strdata)
	var hash_buf bytes.Buffer
	ret := make([]float64, drp.Length)
	drp.data = &ret
	hash := md5.New()
	for i, str := range *strdata {
		switch l := len(str); {
		case l == 0:
			return drp,
				errortypes.NewUserInputError("DRP seems to be invalid. Empty row.")
		case l > 1:
			return drp,
				errortypes.NewUserInputError("DRP seems to be invalid. Too many cols.")
		}
		float, err := strconv.ParseFloat(str[0], 64)
		if err != nil {
			return drp,
				errortypes.NewUserInputError("Row %d: '%s' in drp is not a valid float64", i, str[0])
		}
		ret[i] = float
		binary.Write(&hash_buf, binary.LittleEndian, float)
		if float > drp.Max {
			drp.Max = float
		}
		if float < drp.Min {
			drp.Min = float
		}
		drp.Avg += float
	}
	drp.Avg = drp.Avg / float64(drp.Length)
	if _, err = hash.Write(hash_buf.Bytes()); err != nil {
		return drp, err
	}
	drp.Sha256 = hash.Sum(nil)
	return drp, nil
}
func readCSV(path string) (*[][]string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	csvReader := csv.NewReader(fp)
	csvReader.Comment = '#'
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, errortypes.NewUserInputError(err.Error())
	}
	return &data, nil
}

func readDRPCommentPath(path string, drp *DataRatePattern) (err error) {
	drp.Mapping = make(map[string]string, 5)
	var description strings.Builder
	data, err := os.ReadFile(path)
	if err != nil {
		return errortypes.NewUserInputError("Could not read '%s': %s", path, err)
	}
	var two_values = regexp.MustCompile(`(?m)[0-9]+,[0-9]+`)
	for _, v := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(v, "#") {
			break
		}
		no_white_space := util.RemoveWhiteSpace(v)
		if len(no_white_space) < 2 {
			//Ignore
			continue
		}
		if no_white_space[1] != ':' {
			description.WriteString(strings.TrimSpace(v[1:]) + "\n")
			continue
		}

		sep := strings.Split(no_white_space[2:], "=")
		var_name := sep[0]
		var_value := sep[1]
		if strings.HasPrefix(var_name, "th_") && len(sep) != 2 {
			return fmt.Errorf("error Parsing EvalSetting: '%s'; Not a valid assignment", no_white_space)
		}
		if !two_values.MatchString(var_value) {
			INFO.Printf("Ignoring: %s=%s ", var_name, var_value)
		}
		drp.Mapping[var_name] = fmt.Sprintf("{%s}", var_value)

	}
	drp.Description = description.String()
	return nil
}

func (self *DataRatePatternFileProvider) Provide() (DataRatePattern, error) {
	var ret = DataRatePattern{}
	if self.Path == "" {
		return ret,
			errors.New("DataRatePatternFileProvider was not initialized with a path")
	}
	strdata, err := readCSV(self.Path)
	if err != nil {
		return ret, err
	}
	ret, err = convertDRPdata(strdata)
	if err != nil {
		return ret, err
	}
	err = readDRPCommentPath(self.Path, &ret)
	return ret, err
}
