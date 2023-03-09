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

package persistence

import (
	"errors"
	"reflect"

	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

type Persistence interface {
	Init(login *datatypes.Login) error
	Close() error
	GetSessionStats(session_id int) (load int, start int, end int, err error)
	Persist(obj interface{}) error
	Commit()
	HasDBConnection() bool
	GetStmt() datatypes.SQLStmt
	ValidateUniqueName(obj PersistbleWithUniqueName) error
	ClearCache()
}
type PersistbleWithUniqueName interface {
	ValidateUniqueName(stmt datatypes.SQLStmt) error
}
type DumbPersistable interface {
	Insert(stmt datatypes.SQLStmt) error
	Sync(stmt datatypes.SQLStmt) error
}

// Used for MQ / MP
type BulkPersistable interface {
	GetSQLStatement() string
	GetSQLArgs() []any
}

// Singleton
var (
	persistence_store Persistence = nil
)

func GetPersistence() (*Persistence, error) {
	if persistence_store == nil {
		return nil, errors.New("Persistence was not yet set")
	}
	return &persistence_store, nil
}

// Initially sets the singelton_instance of persistence to v with login
// Subsequent calls have no effect
func SetPersistenceTo(v Persistence, login *datatypes.Login) error {
	if persistence_store != nil {
		WARN.Printf("NOT re-setting persistence from %v to %v\n", persistence_store, v)
		return nil
	}
	DEBUG.Printf("Using %v as Persistence", reflect.TypeOf(v))
	persistence_store = v
	return persistence_store.Init(login)
}
