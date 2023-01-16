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

import (
	"io/fs"
	"sync"
)

type logConfig struct {
	Path     string
	FileName string
	Flag     int
	Perm     fs.FileMode
}

type config struct {
	loggerConfig *logConfig
	player       *DrPlayConfig
	shower       *showConfig
	benchmark    *DrBenchmarkConfig
}

func createNewConfig() (config, error) {
	cfg := config{}
	err := cfg.readFromFile()
	if err != nil {
		return cfg, err
	}
	cfg.setDefaults()
	return cfg, nil
}

var commonConfigMutex sync.Mutex
var (
	commonConfigSingelton *config
)

func getOrCreateCommonConfig() *config {
	if commonConfigSingelton == nil {
		commonConfigMutex.Lock()
		defer commonConfigMutex.Unlock()

		cfg, err := createNewConfig()
		if err != nil {
			panic(err)
		}
		commonConfigSingelton = &cfg

	}
	return commonConfigSingelton
}

func PlayCfg() *DrPlayConfig {
	return getOrCreateCommonConfig().player
}
func ShowCfg() *showConfig {
	return getOrCreateCommonConfig().shower
}
func LoggerCfg() *logConfig {
	return getOrCreateCommonConfig().loggerConfig
}
func BenchmarkCfg() *DrBenchmarkConfig {
	return getOrCreateCommonConfig().benchmark
}
