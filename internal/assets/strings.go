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

package assets

const (
	CFG_FILE_NAME           = "config"
	DRSHOW_EXPORT_PATH_NAME = "export_%s_%s.flow.csv"
)

const (
	URL_BASE_G_MONITORING = "http://%s:3000/d/5-K9sHm4k"
	URL_ARGS_G_MONITORING = "?orgId=1&var-session_id=%d&from=%d&to=%d"
	URL_BASE_G_OVERVIEW   = "http://%s:3000/d/n0rBGsWVz"
	URL_ARGS_G_OVERVIEW   = "?orgId=1&refresh=20s&var-benchmark_id=%d"
)
