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
	"fmt"
	"jens/drcommon/util"
)

type DB_benchmark struct {
	//Name of benchamrk (in json)
	Name string
	//=Tag Name
	Tag string
	//Serial, from DB
	Benchmark_id  int
	Sessions      []*DB_session
	PrintToStdOut bool
	CsvOuptut     bool
	Hash          string
}

func (s *DB_benchmark) GetHashFromLoadedJson() string {
	return s.Hash
}
func (s *DB_benchmark) Insert(stmt SQLStmt) error {
	return stmt.QueryRow(`INSERT INTO benchmark 
	(name, tag) VALUES
	($1, $2) RETURNING benchmark_id`, s.Name, s.Tag).Scan(&s.Benchmark_id)
}
func (s *DB_benchmark) Sync(stmt SQLStmt) error {
	err := stmt.QueryRow(`SELECT benchmark_id FROM benchmark WHERE name=$1 and tag=$2`, s.Name, s.Tag).Scan()
	if err != nil {
		//Todo: Check err type only ErrNoRows
		INFO.Println(err)
		//Case: Does NOT exist
		return s.Insert(stmt)
	} else {
		s.Tag = util.IterateTagName(s.Tag)
		return s.Sync(stmt)
	}
}
func (s *DB_benchmark) DeleteCascade(stmt SQLStmt) error {
	if s.Benchmark_id <= 0 {
		return fmt.Errorf("benchmark id is not set (%d); not deleting", s.Benchmark_id)
	}
	_, err := stmt.Exec("DELETE FROM benchmark WHERE benchmark_id=$1",
		s.Benchmark_id)
	return err
}
func (s *DB_benchmark) GetFkNullable() sql.NullInt64 {
	return sql.NullInt64{Int64: int64(s.Benchmark_id), Valid: s.Benchmark_id > 0}
}
func (s *DB_benchmark) Validate() error {
	for _, v := range s.Sessions {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
}
