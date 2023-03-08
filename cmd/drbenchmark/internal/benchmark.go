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

type Benchmark struct {
	player      *drplay.DrpPlayer
	bm          *datatypes.DB_benchmark
	was_skipped bool
}

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
func New(benchmark *datatypes.DB_benchmark) *Benchmark {
	return &Benchmark{
		player:      nil,
		bm:          benchmark,
		was_skipped: false,
	}
}
func (b *Benchmark) SkipSession() {
	if b.player != nil {
		b.was_skipped = true
		b.player.ExitNoWait()
	}
}
func (b *Benchmark) Play() (err error) {
	/*
	 *    Prep work
	 */
	db, err := persistence.GetPersistence()
	if err != nil {
		return err
	}
	gw := util.RetrieveMostLikelyGatewayIp()
	playtime := 0
	for _, v := range b.bm.Sessions {
		playtime += v.ChildDRP.GetEstimatedPlaytime()
		playtime += 2
	}
	err = (*db).Persist(b.bm)
	if err != nil {
		FATAL.Exit(err)
	}
	fmt.Println("┏╸Starting the benchmark")
	fmt.Printf("┃  ┣╸Name:        %s\n", b.bm.Name)
	fmt.Printf("┃  ┣╸Tag:         %s\n", b.bm.Tag)
	fmt.Printf("┃  ┣╸Hash:        %s\n", b.bm.GetHashFromLoadedJson())
	fmt.Printf("┃  ┣╸Esti. Time:  %ds\n", playtime)
	fmt.Printf("┃  ┗╸BenchmarkID: %d\n", b.bm.Benchmark_id)
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
	session_count := len(b.bm.Sessions)
	for i, v := range b.bm.Sessions {
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
		b.player = drplay.NewDrpPlayer(v)
		err = b.player.Start()
		if err != nil {
			return err
		}
		b.player.Wait()
		(*db).ClearCache()
		_, start_t, end_t, err := (*db).GetSessionStats(v.Session_id)
		if err != nil {
			return err
		}
		if !b.was_skipped {
			fmt.Println(" ✓")
			formatDropDownMessage(1, 2, 2)
			fmt.Printf(assets.URL_BASE_G_MONITORING+assets.URL_ARGS_G_MONITORING+"\n",
				gw, v.Session_id, start_t, end_t)
		} else {
			fmt.Println(" ✖")
			formatDropDownMessage(1, 2, 2)
			fmt.Println("User Interrupt")
			b.was_skipped = false
		}

	}
	/*
	 *    Give combined Summary
	 */
	fmt.Println("┣╸Summarizing benchmark")
	formatDropDownMessage(1, 2, 1)
	fmt.Printf(assets.URL_BASE_G_OVERVIEW+assets.URL_ARGS_G_OVERVIEW+"\n",
		gw, b.bm.Benchmark_id)
	return (*db).Close()
}
