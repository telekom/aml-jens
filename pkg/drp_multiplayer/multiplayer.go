package drpplayer

import "C"
import (
	"fmt"
	"github.com/telekom/aml-jens/internal/config"
	"os"
	"sync"
	"time"

	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
	"github.com/telekom/aml-jens/pkg/drp_player/measuresession"
	"github.com/telekom/aml-jens/pkg/drp_player/trafficcontrol"
)

const MEASURE_FILE_MULTIJENS_PRAEFIX = "/sys/kernel/debug/sch_multijens/0001-"

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

type DrpMultiPlayer struct {
	multisession        *datatypes.DB_multi_session
	session             *datatypes.DB_session
	tc                  *trafficcontrol.TrafficControl
	r                   util.RoutineReport
	is_shutting_down    bool
	close_channel_mutex *sync.Mutex
}

func NewMultiPlayer(config *config.DrPlayConfig) *DrpMultiPlayer {
	var wg sync.WaitGroup
	var close_channel_mutex sync.Mutex
	d := &DrpMultiPlayer{
		is_shutting_down: false,
		multisession:     config.A_MultiSession,
		session:          config.A_Session,
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

func (s *DrpMultiPlayer) Start() error {
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
				if s.is_shutting_down {
					FATAL.Printf("^^ Above exception happend while DrPlay was already shutting down")
				}
				s.ExitNoWait()
			default:
				WARN.Printf("Unknown ErrLevel during Drplay: %+v", err)
			}
			//Something went wrong exit signal to all
			return
		}
	}()

	INFO.Printf("multi play data rate pattern %s on dev %s with %d queues frequency %d hz in loop mode %t\n", s.session.ChildDRP.GetName(), s.session.Dev, s.multisession.UenumTotal, s.session.ChildDRP.Freq, s.session.ChildDRP.IsLooping())
	if err := s.initTC(); err != nil {
		return fmt.Errorf("initTC returned %w", err)
	}
	db, err := persistence.GetPersistence()
	if err != nil {
		FATAL.Println(err)
		os.Exit(4)
	}
	// persist multisession
	err = (*db).Persist(s.multisession)
	if err != nil {
		FATAL.Println(err)
		os.Exit(1)
	}
	for i := 0; i < int(s.multisession.UenumTotal); i++ {
		ueSession := *s.session
		ueSession.Uenum = uint8(i)
		ueSession.ParentMultisession = s.multisession

		err = (*db).Persist(&ueSession)
		if err != nil {
			FATAL.Println(err)
			os.Exit(1)
		}
		// filename to read queue measures
		suffix := fmt.Sprintf("0%d:0", i)
		thisMeasureFilename := MEASURE_FILE_MULTIJENS_PRAEFIX + suffix
		ms := measuresession.NewMeasureSession(&ueSession, s.tc, thisMeasureFilename)
		s.r.Wg.Add(1)
		go ms.Start(s.r)
	}

	s.r.Wg.Add(1)
	go s.tc.LaunchMultiChangeLoop(
		time.Duration(1000/s.session.ChildDRP.Freq)*time.Millisecond,
		s.session.ChildDRP,
		s.r,
	)

	return nil
}
func (s *DrpMultiPlayer) exit_clean() {
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
func (s *DrpMultiPlayer) Wait() {
	s.r.Wg.Wait()
	s.exit_clean()
	DEBUG.Println("Player has exited")
}
func (s *DrpMultiPlayer) close_channel() {
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
func (s *DrpMultiPlayer) ExitNoWait() {
	DEBUG.Println("DrpPlayer was asked to Exit()")
	s.close_channel()
}

func (s *DrpMultiPlayer) Exit() {
	s.ExitNoWait()
	s.Wait()
}
func (s *DrpMultiPlayer) initTC() error {
	s.tc = trafficcontrol.NewTrafficControl(s.session.Dev)
	settings := trafficcontrol.TrafficControlStartParams{
		Datarate:     uint32(s.session.ChildDRP.Peek() * 2),
		QueueSize:    int(s.session.Queuesizepackets),
		AddonLatency: int(s.session.ExtralatencyMs),
		Markfree:     int(s.session.Markfree),
		Markfull:     int(s.session.Markfull),
		Qosmode:      s.session.Qosmode,
		Uenum:        s.multisession.UenumTotal,
	}
	DEBUG.Printf("Init Tc: %+v", settings)
	err := s.tc.InitMultijens(settings,
		trafficcontrol.NftStartParams{
			L4sPremarking:  s.session.L4sEnablePreMarking,
			SignalStart:    s.session.SignalDrpStart,
			DestinationIps: s.multisession.DestinationIps,
			Uenum:          s.multisession.UenumTotal,
		})
	return err
}
