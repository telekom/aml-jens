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
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/internal/util"
	"github.com/telekom/aml-jens/pkg/drp"
)

type DB_data_rate_pattern struct {
	//Set when loading file
	//
	//Pk refrenced in session
	Id           int
	Freq         int
	Scale        float64
	Nomeasure    bool
	MinRateKbits float64
	//Non db-realted
	dr_pattern   drp.DataRatePattern
	path         string
	WarmupTimeMs int32
}

func (s *DB_data_rate_pattern) GetDrp_sha256() []byte {
	return s.dr_pattern.Sha256
}
func (s *DB_data_rate_pattern) GetName() string {
	return s.dr_pattern.Name
}
func (s *DB_data_rate_pattern) GetDescription() sql.NullString {
	return sql.NullString{String: s.dr_pattern.Description, Valid: s.dr_pattern.Description != ""}
}
func (s *DB_data_rate_pattern) GetEstimatedPlaytime() int {
	return int(s.WarmupTimeMs/1000) + (s.dr_pattern.SampleCount() / s.Freq)
}

func (s *DB_data_rate_pattern) GetSQLExistsStatement() string {
	return `select exists (select distinct drp_id from data_rate_pattern where drp_id = $1);`
}
func (s *DB_data_rate_pattern) GetSQLExistsArgs() []any {
	return []any{s.Id}
}

// Returns the Thresholdvalue in the format "{a,b}"
func (s *DB_data_rate_pattern) GetTh_mq_latency() string {
	return s.dr_pattern.Mapping["th_mq_latency"]
}

// Returns the Thresholdvalue in the format "{a,b}"
func (s *DB_data_rate_pattern) GetTh_p95_latency() string {
	return s.dr_pattern.Mapping["th_p95_latency"]
}

// Returns the Thresholdvalue in the format "{a,b}"
func (s *DB_data_rate_pattern) GetTh_p99_latency() string {
	return s.dr_pattern.Mapping["th_p99_latency"]
}

// Returns the Thresholdvalue in the format "{a,b}"
func (s *DB_data_rate_pattern) GetTh_p999_latency() string {
	return s.dr_pattern.Mapping["th_p999_latency"]
}

// Returns the Thresholdvalue in the format "{a,b}"
func (s *DB_data_rate_pattern) GetTh_link_usage() string {
	return s.dr_pattern.Mapping["th_link_usage"]
}
func (s *DB_data_rate_pattern) Insert(stmt SQLStmt) error {
	DEBUG.Printf("Inserting DRP: %v, %v, %v, %v, %v\n", s.dr_pattern.Th_mq_latency,
		s.GetTh_p95_latency(),
		s.GetTh_p99_latency(),
		s.GetTh_p999_latency(),
		s.GetTh_link_usage())
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
		s.GetDrp_sha256(),
		s.GetName(),
		s.GetDescription(),
		s.dr_pattern.Iterator().IsLooping(),
		s.Freq,
		s.Scale,
		s.MinRateKbits,
		s.dr_pattern.Th_mq_latency,
		s.dr_pattern.Th_p95_latency,
		s.dr_pattern.Th_p99_latency,
		s.dr_pattern.Th_p999_latency,
		s.dr_pattern.Th_link_usage).Scan(&s.Id)
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
	if s.dr_pattern.SampleCount() == 0 {
		return errors.New("can't start drplay with a pattern of length 0")
	}
	return nil
}

func (s *DB_data_rate_pattern) GetHash() []byte {
	//This should only be invalid, if no patttern is loaded
	return s.GetDrp_sha256()
}
func (s *DB_data_rate_pattern) GetHashStr() string {
	return hex.EncodeToString(s.GetHash())
}
func (s *DB_data_rate_pattern) SetLooping(endless bool) {
	s.dr_pattern.Iterator().SetLooping(endless)
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

func (drp *DB_data_rate_pattern) ParseDRP(provider drp.DataRatePatternProvider) error {
	if drp.Scale <= 0 {
		return errortypes.NewUserInputError("Scale can't be less than or equal to 0")
	}
	var err error
	drp.dr_pattern, err = provider.Provide(drp.Scale, drp.MinRateKbits)
	return err
}
func (s *DB_data_rate_pattern) GetStats() (min float64, max float64, avg float64) {
	return s.dr_pattern.GetStats()
}
func (s *DB_data_rate_pattern) SetToDone() {
	s.dr_pattern.Iterator().SetDone()
}
func (s *DB_data_rate_pattern) IsLooping() bool {
	return s.dr_pattern.Iterator().IsLooping()
}

// Next returns the next DataRate in a Pattern and its position.
// If there are no Items left (and looping is turned off),
// An error is returned instead
func (drp *DB_data_rate_pattern) Next() (value float64, err error) {
	return drp.dr_pattern.Iterator().Next()
}

// Next returns the next DataRate in a Pattern and its position.
// Does not advance the Iterator
//
//go:inline
func (drp *DB_data_rate_pattern) Peek() (value float64) {
	return drp.dr_pattern.Iterator().Value()
}
func NewDB_data_rate_pattern() *DB_data_rate_pattern {
	return &DB_data_rate_pattern{
		dr_pattern: *drp.NewDataRatePattern("memory"),
	}
}
