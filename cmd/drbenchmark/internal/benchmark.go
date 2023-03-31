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
	playtime *= b.bm.Repetitions + (b.bm.Repetitions * 5)
	err = (*db).Persist(b.bm)
	if err != nil {
		FATAL.Exit(err)
	}
	fmt.Println("┏╸Starting the benchmark")
	fmt.Printf("┃  ┣╸Name:        %s\n", b.bm.Name)
	fmt.Printf("┃  ┣╸Tag:         %s\n", b.bm.Tag)
	fmt.Printf("┃  ┣╸Hash:        %s\n", b.bm.GetHashFromLoadedJson())
	if b.bm.Repetitions > 1 {
		fmt.Printf("┃  ┣╸Repetitions: %d\n", b.bm.Repetitions)
	}
	fmt.Printf("┃  ┣╸Esti. Time:  %ds\n", playtime)
	fmt.Printf("┃  ┗╸BenchmarkID: %d\n", b.bm.Benchmark_id)
	fmt.Println("┣╸Beginning with Sessions")
	session_count := len(b.bm.Sessions)
	for i, v := range b.bm.Sessions {
		session_name_original := v.Name
		for rep := 0; rep < b.bm.Repetitions; rep++ {
			if b.bm.Repetitions > 1 {
				v.Name = fmt.Sprintf("%s [rep%d/%d]", session_name_original, rep+1, b.bm.Repetitions)
				v.ChildDRP.Reset()
			}
			b.pre_play_session(i, session_count, v)
			err, res_session := b.play_session(db, *v)
			if err != nil {
				fmt.Println(" ✖")
				formatDropDownMessage(1, 2, 2)
				fmt.Println("Error while playing session")
				WARN.Println(err)
				continue
			}
			session_id := res_session.Session_id
			b.post_play_session(db, gw, session_id)
			if b.bm.Repetitions > 1 && b.was_skipped {
				time.Sleep(5 * time.Second)
			}
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
func (b *Benchmark) pre_play_session(i int, session_count int, v *datatypes.DB_session) {
	formatDropDownMessage(i, session_count, 1)
	fmt.Printf("'%s' (≈%ds)", v.Name, v.ChildDRP.GetEstimatedPlaytime())
}
func (b *Benchmark) post_play_session(db *persistence.Persistence, gw string, session_id int) {
	_, start_t, end_t, err := (*db).GetSessionStats(session_id)
	if err != nil {
		fmt.Println(" ✖")
		formatDropDownMessage(1, 2, 2)
		fmt.Println("Error while getting session stats")
		WARN.Println(err)
		return
	}
	if !b.was_skipped {
		fmt.Println(" ✓")
		formatDropDownMessage(1, 2, 2)
		fmt.Printf(assets.URL_BASE_G_MONITORING+assets.URL_ARGS_G_MONITORING+"\n",
			gw, session_id, start_t, end_t)
	} else {
		fmt.Println(" ✖")
		formatDropDownMessage(1, 2, 2)
		fmt.Println("User Interrupt")
		b.was_skipped = false
	}
}
func (b *Benchmark) play_session(db *persistence.Persistence, session_copy datatypes.DB_session) (error, *datatypes.DB_session) {
	v := &session_copy
	v.Time = uint64(time.Now().UnixMilli())
	config.PlayCfg().A_Session = v
	if err := (*db).Persist(v); err != nil {
		return err, nil
	}
	v.Time = uint64(time.Now().UnixMilli())
	b.player = drplay.NewDrpPlayer(v)
	if err := b.player.Start(); err != nil {
		return err, nil
	}
	b.player.Wait()
	(*db).ClearCache()

	return nil, v
}
