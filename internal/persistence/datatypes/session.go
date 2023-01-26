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
	"net"

	"github.com/telekom/aml-jens/internal/util"
)

type DB_session struct {
	Session_id          int
	Name                string
	Time                uint64
	Drp_ID              int
	Dev                 string
	Markfree            int32
	Markfull            int32
	Queuesizepackets    int32
	ExtralatencyMs      int32
	L4sEnablePreMarking bool
	Nomeasure           bool
	WarmupTimeMs        int
	//Non DB
	SignalDrpStart bool
	// DB_Relations
	ParentBenchmark *DB_benchmark
	ChildDRP        *DB_data_rate_pattern
	//Todo: Change from persistence map to this!
	knownFlowsByMeasure_ID map[string]int
}

func (s *DB_session) GetFlowIDbyMeasureStr(m string) int {
	res, ok := s.knownFlowsByMeasure_ID[m]
	if ok {
		return res
	}
	return -1
}
func (s *DB_session) AddFlow(f *DB_network_flow) {
	s.knownFlowsByMeasure_ID[f.MeasureIdStr()] = f.Flow_id
}
func (s *DB_session) Insert(stmt SQLStmt) error {
	err := s.ChildDRP.Insert(stmt)
	if err != nil {
		return err
	}

	err = stmt.QueryRow(`INSERT INTO session_tag (
	benchmark_id,
	name,
	time,
	drp_id,
	dev,
	markfree,
	markfull,
	extralatency,
	l4sEnablePreMarking
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING session_id`,
		s.getBenchmarkId(),
		s.Name,
		s.Time,
		s.ChildDRP.Id,
		s.Dev,
		s.Markfree,
		s.Markfull,
		s.ExtralatencyMs,
		s.L4sEnablePreMarking).Scan(&s.Session_id)
	return err
}
func (s *DB_session) Sync(stmt SQLStmt) error {
	DEBUG.Println("Syncing a Session has no effect. Calling Insert instead")
	return s.Insert(stmt)
}

func (s *DB_session) SetParentBenchmark(bm *DB_benchmark) {
	s.ParentBenchmark = bm
}
func (s *DB_session) getBenchmarkId() sql.NullInt64 {
	if s.ParentBenchmark == nil {
		return sql.NullInt64{Int64: 42, Valid: false}
	} else {
		return s.ParentBenchmark.GetFkNullable()
	}
}

func (s *DB_session) ValidateUniqueName(stmt SQLStmt) error {
	exists := false
	err := stmt.QueryRow("select exists (select distinct session_id from session_tag where name = $1 and benchmark_id=$2)",
		s.Name, s.getBenchmarkId()).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		INFO.Printf("Tag with name %s already exists\n", s.Name)
		s.Name = util.IterateTagName(s.Name)
		return s.ValidateUniqueName(stmt)
	}
	return nil
}

func (s *DB_session) Validate() (err error) {
	if _, err := net.InterfaceByName(s.Dev); err != nil {
		FATAL.Exitf("'%s' is not a recognized interface -> %v", s.Dev, err)
	}
	return s.ChildDRP.Validate()
}
