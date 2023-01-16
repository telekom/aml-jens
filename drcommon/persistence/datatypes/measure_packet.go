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

//Implements MassPersistable interface
type DB_measure_packet struct {
	//Only for inprogram use! (=netFlowID)
	Time                uint64
	PacketSojournTimeMs uint32
	LoadKbits           uint32
	Ecn                 uint32
	Dropped             uint32
	Fk_flow_id          int
	Capacitykbits       uint32
}

//go:inline
func (DB_measure_packet) GetSQLStatement() string {
	return "INSERT INTO measure_packet (time, packetsojourntimems, loadkbits, capacitykbits, ecn, dropped, fk_flow_id) VALUES ($1, $2, $3, $4, $5, $6, $7);"
}

//go:inline
func (s *DB_measure_packet) GetSQLArgs() []any {
	return []any{
		s.Time,
		s.PacketSojournTimeMs,
		s.LoadKbits,
		s.Capacitykbits,
		s.Ecn,
		s.Dropped,
		s.Fk_flow_id,
	}
}

//go:inline
func (s *DB_measure_packet) CsvRecord() []string {
	return []string{fmt.Sprint(s.Time), fmt.Sprint(s.PacketSojournTimeMs), fmt.Sprint(s.LoadKbits), fmt.Sprint(s.Capacitykbits), fmt.Sprint(s.Ecn), fmt.Sprint(s.Dropped)}
}

//go:inline
func (s *DB_measure_packet) PrintLine(netflow string) {
	fmt.Println(s.Time, s.PacketSojournTimeMs, s.LoadKbits, s.Capacitykbits, s.Ecn, s.Dropped, netflow)
}
