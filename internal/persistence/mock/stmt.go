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
	"context"
	"database/sql"
	"errors"
)

type SQLStmtMock struct {
}

func (s SQLStmtMock) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MOCK ERROR")
}
func (s SQLStmtMock) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MOCK ERROR")
}
func (s SQLStmtMock) Prepare(query string) (*sql.Stmt, error) {
	return nil, errors.New("MOCK ERROR")
}
func (s SQLStmtMock) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("MOCK ERROR")
}
func (s SQLStmtMock) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MOCK ERROR")
}
func (s SQLStmtMock) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MOCK ERROR")
}
func (s SQLStmtMock) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}
func (s SQLStmtMock) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
