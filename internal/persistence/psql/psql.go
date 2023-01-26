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

package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"

	_ "github.com/lib/pq"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

type DataBase struct {
	db                     *sql.DB
	txMpMutex              sync.Mutex
	txMP                   *sql.Tx
	stmt_packet            *sql.Stmt
	txMQ                   *sql.Tx
	txMqMutex              sync.Mutex
	stmt_queue             *sql.Stmt
	stmt_sessionstats      *sql.Stmt
	knownFlowsByMeasure_ID map[string]int
}

func (s *DataBase) ClearCache() {
	s.knownFlowsByMeasure_ID = make(map[string]int)
}
func (s *DataBase) GetStmt() datatypes.SQLStmt {
	return s.db
}

func (s *DataBase) HasDBConnection() bool {
	return s.db != nil
}
func (s *DataBase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
func (s *DataBase) Init(login *datatypes.Login) error {
	s.knownFlowsByMeasure_ID = make(map[string]int)
	if login != nil {
		db, err := sql.Open("postgres", login.InfoStr())
		if err != nil {
			return fmt.Errorf("could not establish connection to DB: %s", err)
		}
		err = db.Ping()
		if err != nil {
			return err
		}
		s.db = db
		if err != nil {
			return err
		}
		if err = s.prep_bulk_stmts(); err != nil {
			return err
		}
		err = s.prep_special_stmts()
		return err
	}
	return nil
}

func (s *DataBase) prep_bulk_stmts() (err error) {
	s.txMQ, err = s.db.Begin()
	if err != nil {
		return err
	}
	s.txMP, err = s.db.Begin()
	if err != nil {
		return err
	}
	s.stmt_queue, err = s.txMQ.Prepare(datatypes.DB_measure_queue{}.GetSQLStatement())
	if err != nil {
		return err
	}
	s.stmt_packet, err = s.txMP.Prepare(datatypes.DB_measure_packet{}.GetSQLStatement())
	return err
}

func (s *DataBase) prep_special_stmts() (err error) {
	s.stmt_sessionstats, err = s.db.Prepare(`
		select COALESCE(MAX(loadkbits),-1) as load, COALESCE(MIN(time),0) as start, COALESCE(MAX(TIME),1) as end from measure_packet
		where fk_flow_id IN (SELECT flow_id from network_flow where session_id=$1) LIMIT 1;
		`)
	return err
}

func (s *DataBase) GetSessionStats(session_id int) (int, int, int, error) {
	var load = -1
	var start = -1
	var end = -1
	err := s.stmt_sessionstats.QueryRow(session_id).Scan(&load, &start, &end)
	return load, start, end, err
}

// var flow_id_cache map[string]
func (s *DataBase) persist_flow(flow *datatypes.DB_network_flow) error {
	id, keyExists := s.knownFlowsByMeasure_ID[flow.MeasureIdStr()]
	if keyExists {
		flow.Flow_id = id
		return nil
	} else {
		err := flow.Sync(s.db)
		s.knownFlowsByMeasure_ID[flow.MeasureIdStr()] = flow.Flow_id
		return err
	}
}
func (s *DataBase) Persist(obj interface{}) error {
	if !s.HasDBConnection() {
		return errors.New("no connection to db")
	}
	switch v := obj.(type) {
	case datatypes.DB_measure_packet:
		return s.MeasurePacket(v)
	case *datatypes.DB_measure_queue:
		return s.MeasureQueue(v)
	case *datatypes.DB_network_flow:
		if err := s.persist_flow(v); err != nil {
			return fmt.Errorf("Persist(datatypes.DB_network_flow)%v", err)
		}
		return nil
	case persistence.DumbPersistable:
		//Catch for benchmark, data_rate_pattern
		DEBUG.Printf("{interface {persistence.DumbPersistable}} --> %v", reflect.TypeOf(v))
		return v.Insert(s.db)
	default:
		WARN.Printf("unknown Obj-Type: %v (%+v)", reflect.TypeOf(obj), obj)
		return nil
	}
}

//go:inline
func (s *DataBase) MeasurePacket(data datatypes.DB_measure_packet) error {

	if data.Fk_flow_id == -1 {
		FATAL.Exit("Go a MP without flow assigned")
	}
	if data.Capacitykbits == 0 {
		INFO.Println("Not persisting, capacity = 0")
		return nil
		//Do not persist samples where capacity is 0
		//This should only happen during warmup
	}
	_, err := s.stmt_packet.Exec(data.GetSQLArgs()...)
	if err != nil {
		INFO.Printf("MP: %+v -> %v\n", data, err)
	}
	return err
}

//go:inline
func (s *DataBase) MeasureQueue(data *datatypes.DB_measure_queue) error {
	_, err := s.stmt_queue.Exec(data.GetSQLArgs()...)
	return err
}
func (s *DataBase) Commit() {
	go func(self *DataBase) {
		var err error
		self.txMqMutex.Lock()
		//DEBUG.Println("Committing txMQ")
		if err := self.txMQ.Commit(); err != nil {
			FATAL.Exit(err)
		}
		self.txMQ, err = self.db.Begin()
		if err != nil {
			FATAL.Printf("Error while MQ.commit: %s", err)
		}
		self.stmt_queue, err = self.txMQ.Prepare(datatypes.DB_measure_queue{}.GetSQLStatement())
		self.txMqMutex.Unlock()
		if err != nil {
			FATAL.Exit(err)
		}
	}(s)
	go func(self *DataBase) {
		var err error
		self.txMpMutex.Lock()
		//DEBUG.Println("Committing txMP")
		if err := self.txMP.Commit(); err != nil {
			FATAL.Exit(err)
		}
		self.txMP, err = self.db.Begin()
		if err != nil {
			FATAL.Printf("Error while MP.commit: %s", err)
		}
		self.stmt_packet, err = self.txMP.Prepare(datatypes.DB_measure_packet{}.GetSQLStatement())
		self.txMpMutex.Unlock()
		if err != nil {
			FATAL.Exit(err)
		}
	}(s)
}

//go:inline
func (s *DataBase) ValidateUniqueName(obj persistence.PersistbleWithUniqueName) error {
	return obj.ValidateUniqueName(s.db)
}
