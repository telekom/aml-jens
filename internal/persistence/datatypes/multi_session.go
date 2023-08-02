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

type DB_multi_session struct {
	Multisession_id int
	Name            string
	UenumTotal      uint8
	SingleQueue     bool
	FixedNetflows   []string
	UeMinloadkbits  uint32
}

// Executes a sqlstmt to Insert this session into DB
func (s *DB_multi_session) Insert(stmt SQLStmt) error {
	err := stmt.QueryRow(`INSERT INTO multisession (
    name,
    uenum_total
	) VALUES ($1, $2) RETURNING multisession_id`,
		s.Name,
		s.UenumTotal).Scan(&s.Multisession_id)
	return err
}

// == Insert()
func (s *DB_multi_session) Sync(stmt SQLStmt) error {
	WARN.Println("Syncing a Multisession has no effect. Calling Insert instead")
	return s.Insert(stmt)
}
