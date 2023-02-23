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

package drbenchmark

import (
	"fmt"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
	drplay "github.com/telekom/aml-jens/pkg/drp_player"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

func formatDropDownMessage(i int, max int, indent int) {
	p := ""
	switch indent {
	case 1:
		p = "┃  "
	case 2:
		p = "┃  ┃  "
	}

	if i == 0 {
		if max == 1 {
			fmt.Print(p + "┗╸")
		} else {
			fmt.Print(p + "┣╸")
		}
	} else {
		if i == max-1 {
			fmt.Print(p + "┗╸")
		} else {
			fmt.Print(p + "┣╸")
		}
	}
}

func Play(benchmark *datatypes.DB_benchmark) (err error) {
	/*
	 *    Prep work
	 */
	db, err := persistence.GetPersistence()
	if err != nil {
		return err
	}
	gw := util.RetrieveMostLikelyGatewayIp()
	playtime := 0
	for _, v := range benchmark.Sessions {
		playtime += v.ChildDRP.GetEstimatedPlaytime()
		playtime += 2
	}
	err = (*db).Persist(benchmark)
	if err != nil {
		FATAL.Exit(err)
	}
	fmt.Println("┏╸Starting the benchmark")
	fmt.Printf("┃  ┣╸Name:        %s\n", benchmark.Name)
	fmt.Printf("┃  ┣╸Tag:         %s\n", benchmark.Tag)
	fmt.Printf("┃  ┣╸Hash:        %s\n", benchmark.GetHashFromLoadedJson())
	fmt.Printf("┃  ┣╸Esti. Time:  %ds\n", playtime)
	fmt.Printf("┃  ┗╸BenchmarkID: %d\n", benchmark.Benchmark_id)
	/*
		fmt.Printf("┣╸Warmup: Estimating max bitrate (≈%ds)\n", summary.Definition.Inner.MaxBitrateEstimationTimeS)
		if err = summary.Warmup(); err != nil {
			return err
		}
		fmt.Printf("┃  ┗╸Max: %d kb\n", summary.maxBitrateInWarump)
	*/
	/*
	 *    Start playing each session
	 */
	fmt.Println("┣╸Beginning with Sessions")
	session_count := len(benchmark.Sessions)
	for i, v := range benchmark.Sessions {
		v.Time = uint64(time.Now().UnixMilli())
		config.PlayCfg().A_Session = v
		err := (*db).Persist(v)
		if err != nil {
			FATAL.Exit(err)
		}
		formatDropDownMessage(i, session_count, 1)
		fmt.Printf("%s (≈%ds)", v.Name, v.ChildDRP.GetEstimatedPlaytime())

		time.Sleep(2 * time.Second)
		v.Time = uint64(time.Now().UnixMilli())
		player := drplay.NewDrpPlayer(v)
		player.Start()
		(*db).ClearCache()
		_, start_t, end_t, err := (*db).GetSessionStats(v.Session_id)
		if err != nil {
			return err
		}
		fmt.Println(" ✓")
		formatDropDownMessage(1, 2, 2)
		fmt.Printf(assets.URL_BASE_G_MONITORING+assets.URL_ARGS_G_MONITORING+"\n",
			gw, v.Session_id, start_t, end_t)

	}
	/*
	 *    Give combined Summary
	 */
	fmt.Println("┣╸Summarizing benchmark")
	formatDropDownMessage(1, 2, 1)
	fmt.Printf(assets.URL_BASE_G_OVERVIEW+assets.URL_ARGS_G_OVERVIEW+"\n",
		gw, benchmark.Benchmark_id)
	return (*db).Close()
}
