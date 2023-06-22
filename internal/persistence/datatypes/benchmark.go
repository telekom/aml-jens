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

	"github.com/telekom/aml-jens/internal/commands"
	"github.com/telekom/aml-jens/internal/util"
)

type benchmark_callback struct {
	executable string
	ret        chan error
}

type Benchmark_callback_id uint8

const (
	BM_CB_PreBenchmark  = 1
	BM_CB_PreSession    = 4
	BM_CB_PostSession   = 7
	BM_CB_PostBenchmark = 9
)

var callback_name_lookup = []string{
	/*0*/ "None",
	/*1*/ "PreBenchmark",
	/*2*/ "None",
	/*3*/ "None",
	/*4*/ "PreSession",
	/*5*/ "None",
	/*6*/ "None",
	/*7*/ "PostSession",
	/*8*/ "None",
	/*9*/ "PostBenchmark",
}

// Create a new benchmark_callback.
//
// Check existence and executbale status of path
func new_bm_cb(path string) (error, benchmark_callback) {
	var err error
	if path != "" {
		is_exec, err := util.IsFileAndExecutable(path)
		if err != nil {
			path = ""
			err = fmt.Errorf("could not check status fo benchmark_callback %w", err)
		} else if !is_exec {
			path = ""
			err = fmt.Errorf("file %s is not a file or not marked as executable", path)
		}
	}
	return err, benchmark_callback{path, make(chan error)}
}

func (b benchmark_callback) exec(i Benchmark_callback_id, env []string, args ...string) (bool, <-chan error) {
	if b.executable == "" {
		return false, b.ret
	}
	go func() {
		args = append(args, "DUMMY")
		copy(args[1:], args)
		args[0] = fmt.Sprint(i)
		res := commands.ExecCommandEnv(b.executable, env, args...)
		err := res.Error()
		if err != nil {
			err = fmt.Errorf("did not Successfully execute %s -> %w", callback_name_lookup[i], res.Error())
		}
		b.ret <- err
	}()
	return true, b.ret
}
func (b benchmark_callback) OnPreBenchmark(bench *DB_benchmark) (is_running bool, result <-chan error) {
	env := []string{
		fmt.Sprintf("Sessions=%d", len(bench.Sessions)),
	}
	return b.exec(BM_CB_PreBenchmark, env,
		fmt.Sprint(bench.Benchmark_id), bench.Name, bench.Tag)
}
func (b benchmark_callback) OnPreSession(sess_num int) (is_running bool, result <-chan error) {
	return b.exec(BM_CB_PreSession,
		[]string{},
		fmt.Sprint(sess_num),
	)
}
func (b benchmark_callback) OnPostSession(success bool, session_id int, start_t int, end_t int) (is_running bool, result <-chan error) {
	return b.exec(BM_CB_PostSession,
		[]string{},
		fmt.Sprint(map[bool]int8{false: 0, true: 1}[success]),
		fmt.Sprint(session_id),
		fmt.Sprint(start_t),
		fmt.Sprint(end_t))
}
func (b benchmark_callback) OnPostBenchmark(id int) (is_running bool, result <-chan error) {
	return b.exec(BM_CB_PostBenchmark,
		[]string{},
		fmt.Sprint(id))
}

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
	//Not DB
	Callback benchmark_callback
}

// Links external callbacks for benchmark.
// should only be used by argparse
func (s *DB_benchmark) LinkCallback(path string) (err error) {
	err, s.Callback = new_bm_cb(path)
	return err
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
