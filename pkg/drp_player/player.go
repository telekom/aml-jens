package drpplayer

import "C"
import (
	"fmt"
	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/config"
	"os"
	"strconv"
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
const MONITOR_NETFLOW_DURATION_S = 3

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

var sessions []*measuresession.MeasureSession

type DrpPlayer struct {
	multisession        *datatypes.DB_multi_session
	session             *datatypes.DB_session
	tc                  *trafficcontrol.TrafficControl
	r                   util.RoutineReport
	is_shutting_down    bool
	close_channel_mutex *sync.Mutex
}

func NewDrpPlayer(config *config.DrPlayConfig) *DrpPlayer {
	var wg sync.WaitGroup
	var close_channel_mutex sync.Mutex
	d := &DrpPlayer{
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

	INFO.Printf("play data rate pattern %s on dev %s with %d samples/s in loop mode %t\n", s.session.ChildDRP.GetName(), s.session.Dev, s.session.ChildDRP.Freq, s.session.ChildDRP.IsLooping())
	if err := s.initTC(); err != nil {
		return fmt.Errorf("initTC returned %w", err)
	}

	s.tc.ChangeTo(s.session.ChildDRP.Peek() * 1.33)
	select {
	case <-time.After(time.Millisecond * time.Duration(s.session.ChildDRP.WarmupTimeMs)):
	case <-s.r.On_extern_exit_c:
		return nil
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

	if !s.session.Nomeasure {
		// launch default queue UE0
		s.launchMeasureSession(0, db, "", true)
		// launch queues for fixed UEs
		if !s.multisession.SingleQueue {
			for i := 1; i <= len(s.multisession.FixedNetflows); i++ {
				netflowFilter := s.multisession.FixedNetflows[i-1]
				s.launchMeasureSession(i, db, netflowFilter, true)
			}
		}
	}

	// play drp
	s.r.Wg.Add(1)
	go s.tc.LaunchMultiChangeLoop(
		time.Duration(1000/s.session.ChildDRP.Freq)*time.Millisecond,
		s.session.ChildDRP,
		s.r,
	)

	//monitor ue load, synchronize queues
	if !s.multisession.SingleQueue && !s.session.Nomeasure {
		s.r.Wg.Add(1)
		go s.monitorTrafficPerUe(db, s.r)
	}

	return nil
}

func (s *DrpPlayer) launchMeasureSession(i int, db *persistence.Persistence, netflowFilter string, fixedNetflow bool) {
	ueSession := *s.session
	ueSession.Uenum = uint8(i)
	ueSession.NetflowFilter = netflowFilter
	ueSession.ParentMultisession = s.multisession
	ueSession.Name = ueSession.Name + "-UE" + strconv.Itoa(i)
	err := (*db).Persist(&ueSession)
	if err != nil {
		FATAL.Println(err)
		os.Exit(1)
	}
	// filename to read queue measures
	suffix := fmt.Sprintf("0%d:0", i)
	thisMeasureFilename := MEASURE_FILE_MULTIJENS_PRAEFIX + suffix
	ms := measuresession.NewMeasureSession(&ueSession, s.tc, thisMeasureFilename, fixedNetflow)
	s.r.Wg.Add(1)
	go ms.Start(s.r)

	sessions = append(sessions, &ms)
}

func (s *DrpPlayer) monitorTrafficPerUe(db *persistence.Persistence, r util.RoutineReport) {
	sampleDuration := MONITOR_NETFLOW_DURATION_S * time.Second
	monitoringTicker := time.NewTicker(sampleDuration)

	// timer to monitor traffic per UE
	for range monitoringTicker.C {
		INFO.Printf("check ue netflows ...")
		// remove non relevant UEs
		INFO.Printf("remove ues without load...")
		uesChanged := false
		for i := 1; i < len(sessions); i++ {
			ueSession := sessions[i]
			if ueSession.FixedNetflow {
				continue
			}

			var sumLoadBytes uint64 = 0
			for _, aggregated_measure := range ueSession.AggregateMeasurePerNetflow {
				sumLoadBytes += aggregated_measure.SumloadTotalBytes
				aggregated_measure.SumloadTotalBytes = 0
			}
			INFO.Printf("ue%v netflow %v %v byte", ueSession.Session.Uenum, ueSession.Session.NetflowFilter, sumLoadBytes)
			ueLoadKbits := sumLoadBytes / (125 * uint64(sampleDuration.Seconds()))
			if ueLoadKbits < uint64(s.multisession.UeMinloadkbits) {
				ueSession.Wg.Done()
				sessions = append(sessions[:i], sessions[i+1:]...)
				uesChanged = true
				INFO.Printf("ue%v removed", ueSession.Session.Uenum)
			}
		}

		// if traffic in ue0 relevant, launch neu UE
		// get max load in netflow for ue 0
		if uint8(len(sessions)) < s.multisession.UenumTotal {
			ue0Session := sessions[0]
			var maxUe0LoadByte uint64 = 0
			var newNetflow string
			INFO.Printf("ue0")
			for _, aggregated_measure := range ue0Session.AggregateMeasurePerNetflow {
				if aggregated_measure.Net_flow.Source_port != 0 &&
					aggregated_measure.Net_flow.Source_port != 22 &&
					aggregated_measure.Net_flow.Destination_port != 0 &&
					aggregated_measure.Net_flow.Destination_port != 22 {
					INFO.Printf("netflow %v %v byte", aggregated_measure.Net_flow.MeasureIdStr(), aggregated_measure.SumloadTotalBytes)
					if aggregated_measure.SumloadTotalBytes > maxUe0LoadByte {
						maxUe0LoadByte = aggregated_measure.SumloadTotalBytes
						protocolType := aggregated_measure.Net_flow.TransportProtocoll
						newNetflow = fmt.Sprintf("ip saddr %s %s sport %d ip daddr %s %s dport %d",
							aggregated_measure.Net_flow.Source_ip,
							protocolType,
							aggregated_measure.Net_flow.Source_port,
							aggregated_measure.Net_flow.Destination_ip,
							protocolType,
							aggregated_measure.Net_flow.Destination_port)
					}
				}
				aggregated_measure.SumloadTotalBytes = 0
			}
			// add netflow as ue, if relevant
			INFO.Printf("add ue with relevant load...")
			avgLoadKbits := maxUe0LoadByte / (125 * uint64(sampleDuration.Seconds()))
			if avgLoadKbits > uint64(s.multisession.UeMinloadkbits) {
				lastSessionIndex := len(sessions)
				s.launchMeasureSession(lastSessionIndex, db, newNetflow, false)
				uesChanged = true
				INFO.Printf("ue%v netflow %v %v kbits added", lastSessionIndex, newNetflow, avgLoadKbits)
			}
		}

		// reinstanciate marking rules
		if uesChanged {
			INFO.Printf("rebuild nft marking rules")
			trafficcontrol.ResetNFT(assets.NFT_TABLE_UEMARK)
			var netfilter []string
			for _, session := range sessions {
				if session.Session.NetflowFilter != "" {
					netfilter = append(netfilter, session.Session.NetflowFilter)
				}
			}
			err := trafficcontrol.CreateRulesMarkUe(netfilter)
			if err != nil {
				FATAL.Println(err)
			}
		}
	}
}

func (s *DrpPlayer) Exit_clean() {
	if err := s.tc.Close(); err != nil {
		WARN.Printf("Exit: error closing TrafficControl: %+v", err)
	}
	if !s.session.Nomeasure {
		p_ptr, err := persistence.GetPersistence()
		if err != nil {
			WARN.Printf("Exit: Could not get persistence %+v", err)
		}
		(*p_ptr).Commit()
	}
}
func (s *DrpPlayer) Wait() {
	s.r.Wg.Wait()
	s.Exit_clean()
	DEBUG.Println("Player has exited")
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
	s.Wait()
}
func (s *DrpPlayer) initTC() error {
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
			L4sPremarking: s.session.L4sEnablePreMarking,
			SignalStart:   s.session.SignalDrpStart,
			Uenum:         s.multisession.UenumTotal,
			SingleQueue:   s.multisession.SingleQueue,
		},
		s.multisession.FixedNetflows,
	)
	return err
}
