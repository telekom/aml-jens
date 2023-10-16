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
	"strings"
)

type DB_network_flow struct {
	//Serial - from DB
	Flow_id int
	//Fk, refrences DB_session
	TransportProtocoll string
	Session_id         int
	Source_ip          string
	Source_port        uint16
	Destination_ip     string
	Destination_port   uint16
	Prio               uint8
	//Used for caching
	measure_id_str string
}

// Executes a stmt to insert it into the DB
func (s *DB_network_flow) Insert(stmt SQLStmt) error {
	return stmt.QueryRow(`INSERT INTO network_flow 
	(
		session_id,
		source_ip,
		source_port,
		destination_ip,
		destination_port,
	 	transport_protocoll,
	 	prio
	)
	VALUES ( $1, $2, $3, $4, $5, $6, $7)
	RETURNING flow_id;`,
		s.Session_id,
		s.Source_ip,
		s.Source_port,
		s.Destination_ip,
		s.Destination_port,
		s.TransportProtocoll,
		s.Prio,
	).Scan(&s.Flow_id)
}

func (s *DB_network_flow) Update(stmt SQLStmt, flowId int, prio uint8) error {
	_, err := stmt.Exec(`UPDATE network_flow SET prio = $1 WHERE flow_id = $2`, prio, flowId)
	return err
}

// Eyecutes a or multiple sqlstmt, that will make sure its sinked with db
//
// In this case, it makes sure, that its not yet in db.
func (s *DB_network_flow) Sync(stmt SQLStmt) error {
	//DEBUG.Printf("Syncing Flow:%+v", s)
	err := stmt.QueryRow(`SELECT flow_id FROM network_flow 
	WHERE session_id=$1 AND source_ip=$2 AND source_port=$3 AND 
	destination_ip=$4 AND destination_port=$5`,
		s.Session_id,
		s.Source_ip,
		s.Source_port,
		s.Destination_ip,
		s.Destination_port,
	).Scan(&s.Flow_id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			//Case not in db -> insert
			return s.Insert(stmt)
		}
	}
	return err
}

func (s *DB_network_flow) MeasureIdStr() string {
	if s.measure_id_str == "" {
		var builder strings.Builder
		builder.Grow(40)
		fmt.Fprintf(&builder, "%s %s:%d-%s:%d", s.TransportProtocoll, s.Source_ip, s.Source_port, s.Destination_ip, s.Destination_port)
		s.measure_id_str = builder.String()
	}
	return s.measure_id_str
}
