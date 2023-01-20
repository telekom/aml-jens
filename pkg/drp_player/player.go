package drpplayer

import "C"
import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/pkg/drp_player/measuresession"
	"github.com/telekom/aml-jens/pkg/drp_player/trafficcontrol"
)

const CTRL_FILE = "/sys/kernel/debug/sch_janz/0001:v1"

var g_channel_exit = make(chan struct{})
var DEBUG, INFO, FATAL = logging.GetLogger()

func StartDrpPlayer(session *datatypes.DB_session) {
	exit_handler_channel := make(chan os.Signal, 1)
	signal.Notify(exit_handler_channel, syscall.SIGINT, syscall.SIGPIPE, syscall.SIGQUIT)
	logging.LinkExitFunction(func() uint8 {
		exit_handler_channel <- syscall.SIGQUIT
		return 0
	}, 2500)
	INFO.Printf("play data rate pattern %s on dev %s with %d samples/s in loop mode %t\n", session.ChildDRP.GetName(), session.Dev, session.ChildDRP.Freq, session.ChildDRP.IsLooping())

	TC := trafficcontrol.NewTrafficControl(session.Dev)
	time.Sleep(time.Duration(session.ChildDRP.WarmupTimeMs) * time.Millisecond)

	// flag to wait until end of thread
	var wg sync.WaitGroup
	err := TC.Init(trafficcontrol.TrafficControlStartParams{
		Datarate:     uint32(session.ChildDRP.Peek()),
		QueueSize:    int(session.Queuesizepackets),
		AddonLatency: int(session.ExtralatencyMs),
		Markfree:     int(session.Markfree),
		Markfull:     int(session.Markfull),
	},
		trafficcontrol.NftStartParams{
			L4sPremarking: session.L4sEnablePreMarking,
			SignalStart:   session.SignalDrpStart,
		})
	if err != nil {
		FATAL.Exit(err)
	}
	wg.Add(1)
	go TC.LaunchChangeLoop(
		time.Duration(1000/session.ChildDRP.Freq)*time.Millisecond,
		&wg,
		session.ChildDRP,
	)

	DEBUG.Println("... in background")

	// start measure session
	if !session.ChildDRP.Nomeasure {
		go measuresession.Start(session, TC, g_channel_exit)
	}

	done := registerExitHandler(&wg, exit_handler_channel, session, TC)

	wg.Wait()
	exit_handler_channel <- syscall.SIGQUIT
	measuresession.WaitGroup.Wait()
	//Wait for ExitHandler
	<-done
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
