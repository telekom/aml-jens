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
		//Exit listener
		select {
		case msg := <-s.r.Application_has_finished:
			INFO.Println(msg)
			DEBUG.Println("Exiting: Application_has_finished")
			s.Exit()
		case <-s.r.On_extern_exit_c:
			//Send exit signal to all
			INFO.Println("DrPlay was asked to quit")
		case err := <-s.r.Send_error_c:
			WARN.Printf("Something went wrong during DrPlay: %v", err)
			//Something went wrong exit signal to all
			s.errs = append(s.errs, err)
			return
		}
	}()
	logging.LinkExitFunction(func() uint8 {
		s.r.Send_error_c <- fmt.Errorf("FATAL")
		os.Exit(-1)
		return 255
	}, 2500)
	INFO.Printf("play data rate pattern %s on dev %s with %d samples/s in loop mode %t\n", s.session.ChildDRP.GetName(), s.session.Dev, s.session.ChildDRP.Freq, s.session.ChildDRP.IsLooping())

	if err := s.launchTC(); err != nil {
		return err
	}
	// flag to wait until end of thread
	DEBUG.Println("... in background")
	//registerExitHandler(&wg, chans, session, TC)
	// start measure session
	if !s.session.ChildDRP.Nomeasure {
		s.r.Wg.Add(1)
		go measuresession.Start(s.session, s.tc, s.r)
	}

	return nil
}
func (s *DrpPlayer) exit_clean() {
	if err := s.tc.Close(); err != nil {
		WARN.Printf("Exit: error closing TrafficControl: %+v", err)
	}
	if !s.session.ChildDRP.Nomeasure {
		p_ptr, err := persistence.GetPersistence()
		if err != nil {
			WARN.Printf("Exit: Could not get persistence %+v", err)
		}
		(*p_ptr).Commit()
		if err := (*p_ptr).Close(); err != nil {
			WARN.Printf("Exit: Could not close persistence %+v", err)
		}
	}
	fmt.Println(strings.Join(assets.END_OF_DRPLAY[:], " "))
}
func (s *DrpPlayer) Wait() {
	s.r.Wg.Wait()
}
func (s *DrpPlayer) Exit() {
	DEBUG.Println("Exiting")
	close(s.r.On_extern_exit_c)
	DEBUG.Println("Waiting")
	s.Wait()
	DEBUG.Println("exit_clean")
	s.exit_clean()
	DEBUG.Println("Player has exited")
}
func (s *DrpPlayer) launchTC() error {
	s.tc = trafficcontrol.NewTrafficControl(s.session.Dev)
	time.Sleep(time.Duration(s.session.ChildDRP.WarmupTimeMs) * time.Millisecond)
	err := s.tc.Init(trafficcontrol.TrafficControlStartParams{
		Datarate:     uint32(s.session.ChildDRP.Peek()),
		QueueSize:    int(s.session.Queuesizepackets),
		AddonLatency: int(s.session.ExtralatencyMs),
		Markfree:     int(s.session.Markfree),
		Markfull:     int(s.session.Markfull),
	},
		trafficcontrol.NftStartParams{
			L4sPremarking: s.session.L4sEnablePreMarking,
			SignalStart:   s.session.SignalDrpStart,
		})
	if err != nil {
		return err
	}
	s.r.Wg.Add(1)
	go s.tc.LaunchChangeLoop(
		time.Duration(1000/s.session.ChildDRP.Freq)*time.Millisecond,
		s.session.ChildDRP,
		s.r,
	)
	return nil
}
