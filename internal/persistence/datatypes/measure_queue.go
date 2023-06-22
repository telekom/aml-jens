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
	"fmt"
)

type DB_measure_queue struct {
	Time              uint64
	Memoryusagebytes  uint32
	PacketsInQueue    uint16
	CapacityKbits     uint64
	Fk_session_tag_id int
}

//go:inline
func (s DB_measure_queue) GetSQLStatement() string {
	return "INSERT INTO measure_queue (time, memoryusagebytes, packetsinqueue, fk_session_tag_id) VALUES ($1, $2, $3, $4);"
}

//go:inline
func (s *DB_measure_queue) GetSQLArgs() []any {
	return []any{s.Time, s.Memoryusagebytes, s.PacketsInQueue, s.Fk_session_tag_id}
}

//go:inline
func (s *DB_measure_queue) CsvRecord() []string {
	return []string{fmt.Sprint(s.Time), fmt.Sprint(s.Memoryusagebytes), fmt.Sprint(s.PacketsInQueue)}
}
