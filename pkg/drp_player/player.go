package drpplayer

import "C"
import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
	"github.com/telekom/aml-jens/pkg/drp_player/measuresession"
	"github.com/telekom/aml-jens/pkg/drp_player/trafficcontrol"
)

const CTRL_FILE = "/sys/kernel/debug/sch_janz/0001:v1"

var g_channel_exit = make(chan struct{})
var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

type DrpPlayer struct {
	session *datatypes.DB_session
	errs    []error
	tc      *trafficcontrol.TrafficControl
	r       util.RoutineReport
}

func NewDrpPlayer(session *datatypes.DB_session) *DrpPlayer {
	var wg sync.WaitGroup
	d := &DrpPlayer{
		session: session,
		errs:    make([]error, 0, 5),
		r: util.RoutineReport{
			Wg:                       &wg,
			On_extern_exit_c:         make(chan uint8),
			Send_error_c:             make(chan error),
			Application_has_finished: make(chan string),
	},
	}
	return d
}
func (s *DrpPlayer) Start() error {
	go func() {
		err := TC.LaunchChangeLoop(
			time.Duration(1000/session.ChildDRP.Freq)*time.Millisecond,
			&wg,
			session.ChildDRP,
		)
		if err != nil {
			FATAL.Println(err)
		}
	}()

	DEBUG.Println("... in background")
	done := registerExitHandler(&wg, exit_handler_channel, session, TC)
	// start measure session
	if !session.ChildDRP.Nomeasure {
		go measuresession.Start(session, TC, g_channel_exit)
	}
	INFO.Println("[][][][][][][][] Waiting")
	wg.Wait()
	exit_handler_channel <- syscall.SIGQUIT
	INFO.Println("[][][][][][][][] Sending quit")
	measuresession.WaitGroup.Wait()
	INFO.Println("[][][][][][][][] Waiting 2.")
	//Wait for ExitHandler
	<-done
	INFO.Println("[][][][][][][][] DONE")
	INFO.Println(strings.Join(assets.END_OF_DRPLAY[:], " "))
}

// Creates a function that, on SIGINT i.e. Ctrl + c resets the dev using TC
// It does not write to console nor does it return anything
func registerExitHandler(wg *sync.WaitGroup, c chan os.Signal, session *datatypes.DB_session, tc *trafficcontrol.TrafficControl) (done chan uint8) {
	done = make(chan uint8)
	go func() {
		sig := <-c
		INFO.Printf("Received signal: %+v\n", sig)
		session.ChildDRP.SetToDone()
		if !session.ChildDRP.Nomeasure {
			DEBUG.Println("Sending Quit msg")
			go func() { g_channel_exit <- struct{}{} }()
		}

		wg.Wait()
		DEBUG.Println("All threads quit")

		err := tc.Close()
		if err != nil {
			INFO.Printf("Error closting TrafficControl: %+v", err)
		}
		measuresession.WaitGroup.Wait()
		if !session.ChildDRP.Nomeasure {
			p_ptr, err := persistence.GetPersistence()
			if err != nil {
				FATAL.Exit(err)
			}
			(*p_ptr).Commit()
		}

		//TODO: On exit write all db objects to db
		DEBUG.Println("Program->Exit")
		done <- 0
	}()
	return done
}
