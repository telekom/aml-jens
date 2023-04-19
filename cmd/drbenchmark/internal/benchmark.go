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
	w := util.NewIndentWirter()
	db, err := persistence.GetPersistence()
	if err != nil {
		return err
	}
	gw := util.RetrieveMostLikelyGatewayIp()
	playtime := 0
	for _, v := range b.bm.Sessions {
		playtime += v.ChildDRP.GetEstimatedPlaytime()
	}
	err = (*db).Persist(b.bm)
	if err != nil {
		FATAL.Exit(err)
	}
	w.WriteNoIndent("DRBENCHMARK\n")
	w.Indent(true)
	w.WriteNormalLines([]string{
		fmt.Sprintf("Name:        %s\n", b.bm.Name),
		fmt.Sprintf("Tag:         %s\n", b.bm.Tag),
		fmt.Sprintf("Hash:        %s\n", b.bm.GetHashFromLoadedJson())})
	w.WriteNormal(fmt.Sprintf(
		"Esti. Time:  %ds\n", playtime))
	w.WriteCloseIndent(fmt.Sprintf(
		"BenchmarkID: %d\n", b.bm.Benchmark_id))
	w.UnIndent()
	w.WriteNormal("Beginning with Benchmarking\n")
	w.Indent(true)
	if is_running, ret := b.bm.Callback.OnPreBenchmark(b.bm); is_running {
		w.Indent(true)
		w.WriteCloseIndent("OnPreBenchmark")
		if err := <-ret; err != nil {
			WARN.Printf("OnPreBenchmark: %s", err)
			w.WriteNoIndent(" ✖\n")
			FATAL.Exit("OnPreBenchmark= something went wrong while calling the script\n")
		}
		w.WriteNoIndent(" ✓\n")
		w.UnIndent()
	}
	session_count := len(b.bm.Sessions)
	for i, v := range b.bm.Sessions {
		msg := fmt.Sprintf(
			"'%s' (≈%ds)\n", v.Name, v.ChildDRP.GetEstimatedPlaytime())
		if i != session_count-1 {
			w.WriteNormal(msg)
		} else {
			w.WriteCloseIndent(msg)
		}
		w.Indent(i != session_count-1)
		is_running, result_c := b.bm.Callback.OnPreSession(i)
		if is_running {
			w.WriteNormal("OnPreSession ")
			err := <-result_c
			if err != nil {
				w.WriteNoIndent(" ✖\n")
				WARN.Println(err)
			} else {
				w.WriteNoIndent(" ✓\n")
			}

		}
		w.WriteNormal("DrPlay      ")
		err, res_session := b.play_session(db, *v)
		if err != nil {
			w.WriteNoIndent(" ✖\n")
			w.Indent(i != session_count-1)
			w.WriteNormal("Error while playing session\n")
			w.WriteCloseIndent(err.Error())
			w.UnIndent()
			WARN.Println(err)
			continue
		}
		session_id := res_session.Session_id
		b.post_play_session(w, db, gw, session_id)
	}
	/*
	 *    Give combined Summary
	 */
	w.UnIndent()
	w.WriteCloseIndent("Summarizing benchmark\n")
	w.Indent(false)
	if !b.was_skipped {
		if is_running, ret := b.bm.Callback.OnPostBenchmark(b.bm.Benchmark_id); is_running {
			w.WriteNormal("OnPostBenchmark")
			if err := <-ret; err != nil {
				w.WriteNoIndent(" ✖\n")
				w.Indent(false)
				w.WriteCloseIndent(err.Error())
				WARN.Printf("OnPostBenchmark: %s\n", err)
			} else {
				w.WriteNoIndent(" ✓\n")
			}
		}
	}

	w.WriteCloseIndent(fmt.Sprintf(
		assets.URL_BASE_G_OVERVIEW+assets.URL_ARGS_G_OVERVIEW+"\n",
		gw, b.bm.Benchmark_id))
	w.UnIndent()
	return (*db).Close()
}

func (b *Benchmark) post_play_session(w *util.IndentedWriter, db *persistence.Persistence, gw string, session_id int) {
	_, start_t, end_t, err := (*db).GetSessionStats(session_id)
	if err != nil {
		w.WriteNoIndent(" ✖\n")
		w.Indent(false)
		w.WriteCloseIndent("Error while getting session stats\n")
		w.UnIndent()
		WARN.Println(err)
		return
	}

	if !b.was_skipped {
		fmt.Println(" ✓")

		if is_running, ret := b.bm.Callback.OnPostSession(true, session_id, start_t, end_t); is_running {

			w.WriteNormal("OnPostSession")
			if err := <-ret; err != nil {
				w.WriteNoIndent(" ✖\n")
				WARN.Println(err)
			} else {
				w.WriteNoIndent(" ✓\n")
			}

		}
		w.WriteCloseIndent(fmt.Sprintf(
			assets.URL_BASE_G_MONITORING+assets.URL_ARGS_G_MONITORING+"\n",
			gw, session_id, start_t, end_t))
		w.UnIndent()
		time.Sleep(250 * time.Millisecond)
	} else {
		w.WriteNoIndent(" ✖\n")
		w.Indent(false)
		w.WriteCloseIndent("User Interrupt\n")
		w.UnIndent()
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
