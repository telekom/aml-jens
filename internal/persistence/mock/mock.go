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

package mock

import (
	"errors"
	"time"

	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
)

var DEBUG, _, _, _ = logging.GetLogger()

type Database struct {
	knownFlowsByMeasure_ID map[string]int
}

func (s *Database) ClearCache() {
	s.knownFlowsByMeasure_ID = make(map[string]int)
}

func (s *Database) print(txt string, args ...any) {
	//DEBUG.Printf(txt, args...)
}
func (s *Database) Close() error {
	s.print("Closing Persistence")
	return nil
}
func (s *Database) GetSessionStats(i int) (int, int, int, error) {
	s.print("Gettting fake SessionStats")
	t := time.Now().Unix()
	return 99999, int(t - 100), int(t), nil
}

func (s *Database) Persist(obj interface{}) error {
	s.print("Persisting: %+v", obj)
	return nil
}

func (s *Database) Commit() {
	s.print("Committing")
}
func (s *Database) Init(login *datatypes.Login) error {
	s.knownFlowsByMeasure_ID = make(map[string]int)
	if login != nil {
		s.print("Initialized PersistenceMock")
		return nil
	}
	return errors.New("no login supplied")
}

func (s *Database) InitTransactions() error {
	return nil
}

func (s *Database) HasDBConnection() bool {
	return false
}
func (s *Database) GetStmt() datatypes.SQLStmt {
	return SQLStmtMock{}
}
func (s *Database) ValidateUniqueName(obj persistence.PersistbleWithUniqueName) error {
	return nil
}
