package drpplayer

import "C"
import (
	"fmt"
	"sync"
	"time"

	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
	"github.com/telekom/aml-jens/pkg/drp_player/measuresession"
	"github.com/telekom/aml-jens/pkg/drp_player/trafficcontrol"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

type DrpPlayer struct {
	session             *datatypes.DB_session
	tc                  *trafficcontrol.TrafficControl
	r                   util.RoutineReport
	is_shutting_down    bool
	close_channel_mutex *sync.Mutex
}

func NewDrpPlayer(session *datatypes.DB_session) *DrpPlayer {
	var wg sync.WaitGroup
	var close_channel_mutex sync.Mutex
	d := &DrpPlayer{
		is_shutting_down: false,
		session:          session,
		r: util.RoutineReport{
			Wg:               &wg,
			On_extern_exit_c: make(chan uint8),
			Send_error_c: make(chan struct {
				Err   error
				Level util.ErrorLevel
			}),
			Application_has_finished: make(chan string),
		},
		close_channel_mutex: &close_channel_mutex,
	}
	return d
}

// Starts every component needed for the operation of Drplayer
//
// Includes, but is not limited to:
//
// - Persistence
//   - csv &/ psql
//
// - Measure session
//
// - Data aggregation
//
// Non blocking
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
			switch err.Level {
			case util.ErrInfo:
				INFO.Printf("During DrPlay: %v", err.Err)
			case util.ErrWarn:
				WARN.Printf("During DrPlay: %v", err.Err)
			case util.ErrFatal:
				FATAL.Printf("Something went wrong during DrPlay: %v", err.Err)
				s.close_channel()
				if s.is_shutting_down {
					FATAL.Printf("^^ Above exception happend while DrPlay was already shutting down")
				}
			default:
				WARN.Printf("Unknown ErrLevel during Drplay: %+v", err)
			}
			//Something went wrong exit signal to all
			return
		}
	}()

	INFO.Printf("play data rate pattern %s on dev %s with %d samples/s in loop mode %t\n", s.session.ChildDRP.GetName(), s.session.Dev, s.session.ChildDRP.Freq, s.session.ChildDRP.IsLooping())

	if err := s.initTC(); err != nil {
		return fmt.Errorf("initTC returned %w", err)
	}

	if !s.session.ChildDRP.Nomeasure {
		ms := measuresession.NewMeasureSession(s.session, s.tc)
		s.r.Wg.Add(1)
		go ms.Start(s.r)
	}
	s.r.Wg.Add(1)
	go s.tc.LaunchChangeLoop(
		time.Duration(1000/s.session.ChildDRP.Freq)*time.Millisecond,
		s.session.ChildDRP,
		s.r,
	)
	return nil
}
func (s *DrpPlayer) exit_clean() {
	DEBUG.Println("exit_clean")
	if err := s.tc.Close(); err != nil {
		WARN.Printf("Exit: error closing TrafficControl: %+v", err)
	}
	if !s.session.ChildDRP.Nomeasure {
		p_ptr, err := persistence.GetPersistence()
		if err != nil {
			WARN.Printf("Exit: Could not get persistence %+v", err)
		}
		(*p_ptr).Commit()
	}
}
func (s *DrpPlayer) Wait() {
	s.r.Wg.Wait()
	s.exit_clean()
}
func (s *DrpPlayer) close_channel() {
	s.close_channel_mutex.Lock()
	if s.is_shutting_down {
		s.close_channel_mutex.Unlock()
		DEBUG.Println("Drplay is already closing itself")
		return
	}
	DEBUG.Println("Closing channel")
	close(s.r.On_extern_exit_c)
	s.is_shutting_down = true
	s.close_channel_mutex.Unlock()

}
func (s *DrpPlayer) ExitNoWait() {
	DEBUG.Println("DrpPlayer was asked to Exit()")
	s.close_channel()
}

func (s *DrpPlayer) Exit() {
	s.ExitNoWait()
	DEBUG.Println("Waiting for routines to end")
	s.Wait()
	DEBUG.Println("Player has exited")
}
func (s *DrpPlayer) initTC() error {
	s.tc = trafficcontrol.NewTrafficControl(s.session.Dev)
	//Resetting qdisc
	//might fail - ignore
	select {
	case <-s.r.On_extern_exit_c:
		return nil
	case <-time.After(time.Duration(s.session.ChildDRP.WarmupTimeMs) * time.Millisecond):
		_ = s.tc.Reset()
	}
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
	return err
}
