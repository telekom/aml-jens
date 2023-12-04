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
	"encoding/hex"
	"errors"

	"github.com/telekom/aml-jens/internal/errortypes"
	"github.com/telekom/aml-jens/pkg/drp"
)

type DB_data_rate_pattern struct {
	//Set when loading file
	//
	//Pk refrenced in session
	Id        int
	Freq      int
	Nomeasure bool
	//Non db-realted
	dr_pattern     drp.DataRatePattern
	WarmupTimeMs   int32
	RawMeasureDump bool

	Intial_minRateKbits float64
	Initial_scale       float64
}

// Return the name of the loaded pattern; most likely filename
//
//go:inline
func (s *DB_data_rate_pattern) GetName() string {
	return s.dr_pattern.Name
}

// Return the description of the loaded pattern; most likely empty
//
// The description is any comment in a dpr.csv that is not in the :key=value format
//
//go:inline
func (s *DB_data_rate_pattern) GetDescription() sql.NullString {
	return sql.NullString{String: s.dr_pattern.Description, Valid: s.dr_pattern.Description != ""}
}

// Returns the playtime, extrapolated from frequency, samples and warmup
//
//go:inline
func (s *DB_data_rate_pattern) GetEstimatedPlaytime() int {
	return int(s.WarmupTimeMs/1000) + (s.dr_pattern.SampleCount() / s.Freq)
}

//go:inline
func (s *DB_data_rate_pattern) GetSQLExistsStatement() string {
	return `select exists (select distinct drp_id from data_rate_pattern where drp_id = $1);`
}

//go:inline
func (s *DB_data_rate_pattern) GetSQLExistsArgs() []any {
	return []any{s.Id}
}

// Returns the Thresholdvalue in the format "{a,b}"
//
//go:inline
func (s *DB_data_rate_pattern) GetTh_mq_latency() string {
	return s.dr_pattern.GetMappingValue("th_mq_latency", "{}")
}

// Returns the Thresholdvalue in the format "{a,b}"
//
//go:inline
func (s *DB_data_rate_pattern) GetTh_p95_latency() string {
	return s.dr_pattern.GetMappingValue("th_p95_latency", "{}")
}

// Returns the Thresholdvalue in the format "{a,b}"
//
//go:inline
func (s *DB_data_rate_pattern) GetTh_p99_latency() string {
	return s.dr_pattern.GetMappingValue("th_p99_latency", "{}")
}

// Returns the Thresholdvalue in the format "{a,b}"
//
//go:inline
func (s *DB_data_rate_pattern) GetTh_p999_latency() string {
	return s.dr_pattern.GetMappingValue("th_p999_latency", "{}")
}

// Returns the Thresholdvalue in the format "{a,b}"
//
//go:inline
func (s *DB_data_rate_pattern) GetTh_link_usage() string {
	return s.dr_pattern.GetMappingValue("th_link_usage", "{}")
}
func (s *DB_data_rate_pattern) Insert(stmt SQLStmt) error {
	DEBUG.Printf("Inserting DRP: %v, %v, %v, %v, %v\n", s.GetTh_mq_latency(),
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
		s.GetHash(),
		s.GetName(),
		s.GetDescription(),
		s.dr_pattern.Iterator().IsLooping(),
		s.Freq,
		s.dr_pattern.GetScale(),
		s.dr_pattern.GetMinRateKbits(),
		s.GetTh_mq_latency(),
		s.GetTh_p95_latency(),
		s.GetTh_p99_latency(),
		s.GetTh_p999_latency(),
		s.GetTh_link_usage()).Scan(&s.Id)
	return err
}
func (s *DB_data_rate_pattern) Sync(stmt SQLStmt) error {
	WARN.Println("Syncing a DRP has no effect. INSERTING instead")
	return s.Insert(stmt)
}

func (s *DB_data_rate_pattern) Validate() (err error) {
	if s.Freq < 1 || s.Freq > 100 {
		return errortypes.NewUserInputError("frequency must be in intervall [1..100]")
	}

	if s.dr_pattern.GetScale() < 0.1 {
		return errortypes.NewUserInputError("scale factor must be greater than 0.1")
	}
	if s.dr_pattern.SampleCount() == 0 {
		return errors.New("can't start drplay with a pattern of length 0")
	}
	return nil
}

// Returns the sha256 byte array of the loaded pattern
//
//go:inline
func (s *DB_data_rate_pattern) GetHash() []byte {
	//This should only be invalid, if no patttern is loaded
	return s.dr_pattern.Sha256
}

// Returns the sha256 byte array as a hex string of the loaded pattern
//
//go:inline
func (s *DB_data_rate_pattern) GetHashStr() string {
	return hex.EncodeToString(s.GetHash())
}

// Set Looping mode to <endless>
//
// true : looping
//
// false: once
//
//go:inline
func (s *DB_data_rate_pattern) SetLooping(endless bool) {
	s.dr_pattern.Iterator().SetLooping(endless)
}

// Load a DataRatePattern from the given Provider.
// Uses the self.Internal_{scale, minRateKbits} as constraints to the provider.
func (drp *DB_data_rate_pattern) ParseDRP(provider drp.DataRatePatternProvider) error {
	if drp.Initial_scale <= 0 {
		return errortypes.NewUserInputError("Scale can't be less than or equal to 0")
	}
	var err error
	drp.dr_pattern, err = provider.Provide(drp.Initial_scale, drp.Intial_minRateKbits)
	return err
}

// Wraps drp.DataRatePattern{}.GetStats()
//
//go:inline
func (s *DB_data_rate_pattern) GetStats() (min float64, max float64, avg float64) {
	return s.dr_pattern.GetStats()
}

// Wraps drp.DataRatePattern{}.SetToDone()
//
//go:inline
func (s *DB_data_rate_pattern) SetToDone() {
	s.dr_pattern.Iterator().SetDone()
}

// Wraps drp.DataRatePattern{}.Iterator().SetToDone()
//
//go:inline
func (s *DB_data_rate_pattern) IsLooping() bool {
	return s.dr_pattern.Iterator().IsLooping()
}

// Next returns the next DataRate in a Pattern and its position.
// If there are no Items left (and looping is turned off),
// An error is returned instead
//
// Wraps drp.DataRatePattern{}.Iterator().Next()
//
//go:inline
func (drp *DB_data_rate_pattern) Next() (value float64, err error) {
	return drp.dr_pattern.Iterator().Next()
}

// Next returns the next DataRate in a Pattern and its position.
// Does not advance the Iterator
//
// Wraps drp.DataRatePattern{}.Iterator().Value()
//
//go:inline
func (drp *DB_data_rate_pattern) Peek() (value float64) {
	return drp.dr_pattern.Iterator().Value()
}

// Create a new DataBaseObject with some initalized values
func NewDB_data_rate_pattern() *DB_data_rate_pattern {
	return &DB_data_rate_pattern{
		Initial_scale:       1,
		Intial_minRateKbits: 0,
		dr_pattern: *drp.NewDataRatePattern(struct {
			MinRateKbits float64
			Scale        float64
			Origin       string
		}{MinRateKbits: 0, Scale: 1, Origin: "memory"}),
	}
}
