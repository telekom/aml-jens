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

package config

import "jens/drcommon/persistence/datatypes"

type DrPlayConfig struct {
	Psql datatypes.Login
	//The following will be set by drplayer:
	PrintToStdOut bool
	A_Session     *datatypes.DB_session
	A_Csv         bool
	A_Benchmark   *datatypes.DB_benchmark
}

func (s DrPlayConfig) GetDBObj() datatypes.DB_session {
	//s.A_Session.ChildDRP.Th_link_usage = s.OLD_Pattern.Evaluation.LinkUsage
	//s.A_Session.ChildDRP.Th_mq_latency = s.OLD_Pattern.Evaluation.MQLatency
	//s.A_Session.ChildDRP.Th_p95_latency = s.OLD_Pattern.Evaluation.P95Latency
	//s.A_Session.ChildDRP.Th_p99_latency = s.OLD_Pattern.Evaluation.P99Latency
	//s.A_Session.ChildDRP.Th_p999_latency = s.OLD_Pattern.Evaluation.P999Latency
	//s.A_Session.ChildDRP.Description = sql.NullString{s.OLD_Pattern.Description, s.OLD_Pattern.Description != ""}
	return *s.A_Session

}
