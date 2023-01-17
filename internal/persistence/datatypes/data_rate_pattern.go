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

package datatypes

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/util"
)

type DB_data_rate_pattern struct {
	//Set when loading file
	//
	//Pk refrenced in session
	Id         int
	Drp_sha256 []byte
	//Set when loading file
	Name string
	//Set when loading file
	Description     sql.NullString
	Loop            bool
	Freq            int
	Scale           float64
	Nomeasure       bool
	MinRateKbits    float64
	Th_mq_latency   string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_p95_latency  string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_p99_latency  string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_p999_latency string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	Th_link_usage   string //Format: '{a,b}' | a,b ∈ [0-9]+ //[2]float64
	//Non db-realted
	path         string
	pattern      []float64
	position     int
	operator     int8
	WarmupTimeMs int32
	//Stats
	min float64
	max float64
	avg float64
}

func (s *DB_data_rate_pattern) GetEstimatedPlaytime() int {
	return int(s.WarmupTimeMs/1000) + (len(s.pattern) / s.Freq)
}

func (s *DB_data_rate_pattern) GetSQLExistsStatement() string {
	return `select exists (select distinct drp_id from data_rate_pattern where drp_id = $1);`
}
func (s *DB_data_rate_pattern) GetSQLExistsArgs() []any {
	return []any{s.Id}
}

func (s *DB_data_rate_pattern) Insert(stmt SQLStmt) error {
	DEBUG.Printf("Inserting DRP: %v, %v, %v, %v, %v\n", s.Th_mq_latency,
		s.Th_p95_latency,
		s.Th_p99_latency,
		s.Th_p999_latency,
		s.Th_link_usage)
	err := stmt.QueryRow(`INSERT INTO data_rate_pattern
	(
		drp_sha256,
		"name",
		description,
		loop,
		freq,
		scale,
		minrateKbits,
		th_mq_latency,
		th_p95_latency,
		th_p99_latency,
		th_p999_latency,
		th_link_usage
	)
	VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12 )
	RETURNING drp_id;`,
		s.Drp_sha256,
		s.Name,
		s.Description,
		s.Loop,
		s.Freq,
		s.Scale,
		s.MinRateKbits,
		s.Th_mq_latency,
		s.Th_p95_latency,
		s.Th_p99_latency,
		s.Th_p999_latency,
		s.Th_link_usage).Scan(&s.Id)
	return err
}
func (s *DB_data_rate_pattern) Sync(stmt SQLStmt) error {
	FATAL.Println("Syncing a DRP has no effect. INSERTING instead")
	return s.Insert(stmt)
}

func (s *DB_data_rate_pattern) Validate() (err error) {
	if s.Freq < 1 || s.Freq > 100 {
		return errortypes.NewUserInputError("frequency must be in intervall [1..100]")
	}

	if s.Scale < 0.1 {
		return errortypes.NewUserInputError("scale factor must be greater than 0.1")
	}
	if s.pattern == nil {
		return errors.New("dB_data_rate_pattern.pattern is nil")
	}
	if len(s.pattern) == 0 {
		return errors.New("can't start drplay with a pattern of length 0")
	}
	return nil
}

func (s *DB_data_rate_pattern) GetHash() []byte {
	//This should only be invalid, if no patttern is loaded
	return s.Drp_sha256
}
func (s *DB_data_rate_pattern) GetHashStr() string {
	return hex.EncodeToString(s.GetHash())
}
func (s *DB_data_rate_pattern) SetLooping(endless bool) {
	s.Loop = endless
}
func readCSV(path string) (data [][]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errortypes.NewUserInputError("Could not read '%s': %s", path, err)
	}
	// remember to close the file at the end of the program
	defer f.Close()
	// read csv values
	csvReader := csv.NewReader(f)
	csvReader.Comment = '#'
	data, err = csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readDRPCommentPath(path string) (comment string, settings map[string]string, err error) {
	settings = make(map[string]string, 5)
	var description strings.Builder
	data, err := os.ReadFile(path)
	if err != nil {
		return "", settings, errortypes.NewUserInputError("Could not read '%s': %s", path, err)
	}
	var two_values = regexp.MustCompile(`(?m)[0-9]+,[0-9]+`)
	for _, v := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(v, "#") {
			return description.String(), settings, nil
		}
		no_white_space := util.RemoveWhiteSpace(v)
		if len(no_white_space) < 2 {
			//Ignore
			continue
		}
		if no_white_space[1] != ':' {
			description.WriteString(v[1:] + "\n")
			continue
		}

		E := func(s string) error { return fmt.Errorf("error Parsing EvalSetting: '%s'; %s", no_white_space, s) }
		sep := strings.Split(no_white_space[2:], "=")
		var_name := sep[0]
		var_value := sep[1]
		if len(sep) != 2 {
			return "", settings, E("Not a valid assignment")
		}
		if !two_values.MatchString(var_value) {
			INFO.Printf("Ignoring: %s=%s ", var_name, var_value)
		}
		settings[var_name] = fmt.Sprintf("{%s}", var_value)

	}
	return description.String(), settings, nil
}

// Loads a DRP file (*.csv) from the filesystem
// Applys scale and minrate on load. Sets looping.
// Changes *session
func (drp *DB_data_rate_pattern) ParseDrpFile(path string) error {
	var hash_buf bytes.Buffer
	drp.operator = 1
	drp.path = path
	split := strings.Split(path, "/")
	drp.Name = split[len(split)-1]

	comment, dpr_eval_setting, err := readDRPCommentPath(path)
	if err != nil {
		return err
	}
	drp.Description = sql.NullString{String: comment, Valid: comment != ""}
	value, ok := dpr_eval_setting["th_link_usage"]
	if ok {
		drp.Th_link_usage = value
	}
	value, ok = dpr_eval_setting["th_mq_latency"]
	if ok {
		drp.Th_mq_latency = value
	}
	value, ok = dpr_eval_setting["th_p95_latency"]
	if ok {
		drp.Th_p95_latency = value
	}
	value, ok = dpr_eval_setting["th_p99_latency"]
	if ok {
		drp.Th_p99_latency = value
	}
	value, ok = dpr_eval_setting["th_p999_latency"]
	if ok {
		drp.Th_p999_latency = value
	}
	data, err := readCSV(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errortypes.NewUserInputError("DRP @ '%s' seems to be invalid. No data loaded.", path)
	}
	if len(data[0]) != 1 {
		return errortypes.NewUserInputError("DRP @ '%s' seems to be invalid. Too many cols loaded.", path)
	}

	drp.pattern = make([]float64, len(data))
	drp.max = -1
	drp.min = math.MaxFloat64
	drp.avg = 0
	for i, v := range data {
		float, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			fmt.Printf("Error while Reading DRP: %s\n", path)
			return fmt.Errorf("DRP (%s) in line %d '%s' : %s", path, i, v, err.Error())
		}
		drp.pattern[i] = math.Max(float*drp.Scale, drp.MinRateKbits)
		binary.Write(&hash_buf, binary.LittleEndian, float)
		if float > drp.max {
			drp.max = float
		}
		if float < drp.min {
			drp.min = float
		}
		drp.avg += float
	}
	drp.avg = drp.avg / float64(len(data)-1)
	hash := md5.New()
	_, err = hash.Write(hash_buf.Bytes())
	if err != nil {
		return err
	}
	drp.Drp_sha256 = hash.Sum(nil)
	return nil
}
func (s *DB_data_rate_pattern) GetStats() (min float64, max float64, avg float64) {
	return s.min, s.max, s.avg
}
func (s *DB_data_rate_pattern) GetInternalPattern() []float64 {
	return s.pattern
}
func (s *DB_data_rate_pattern) SetToDone() {
	s.Loop = false
	s.operator = +1
	s.position = len(s.pattern)
}
func (s *DB_data_rate_pattern) IsLooping() bool {
	return s.Loop
}

// Next returns the next DataRate in a Pattern and its position.
// If there are no Items left (and looping is turned off),
// An error is returned instead
func (drp *DB_data_rate_pattern) Next() (value float64, err error) {
	last_pos := drp.position
	if drp.position >= len(drp.pattern) {
		if drp.Loop {
			drp.operator = -1
			last_pos -= 1
			drp.position -= 1
		} else {
			return -1, &errortypes.IterableStopError{Msg: "end of DRP"}
		}
	}
	if drp.position < 0 && drp.Loop {
		drp.operator = +1
		last_pos += 1
		drp.position += 1
	}

	drp.position += int(drp.operator)
	return drp.pattern[last_pos], nil
}

// Next returns the next DataRate in a Pattern and its position.
// Does not advance the Iterator
//
//go:inline
func (drp *DB_data_rate_pattern) Peek() (value float64) {
	return drp.pattern[drp.position]
}
func NewDB_data_rate_pattern() *DB_data_rate_pattern {
	return &DB_data_rate_pattern{
		Th_mq_latency:   "{2,4}",
		Th_p95_latency:  "{10,20}",
		Th_p99_latency:  "{10,20}",
		Th_p999_latency: "{10,20}",
		Th_link_usage:   "{60,80}",
	}
}
