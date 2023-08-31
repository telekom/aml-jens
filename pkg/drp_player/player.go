package drpplayer

import "C"
import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/config"
	"github.com/telekom/aml-jens/internal/errortypes"

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
			Wg:              &wg,
			Exit_now_signal: make(chan uint8),
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

// Starts a goroutine that will change the current capacity restricitons.
// A change will occur after the waitTime is exceeded.
//
// # Uses util.RoutineReport
//
// Blockig - also spawns 1 short lived routine
func (s *DrpPlayer) launchMultiChangeLoop(waitTime time.Duration, drp *datatypes.DB_data_rate_pattern, r util.RoutineReport, shareCapacityResources bool) {
	ticker := time.NewTicker(waitTime)
	INFO.Printf("start playing DataRatePattern @%s", waitTime.String())
	for {
		select {
		case <-r.Exit_now_signal:
			DEBUG.Println("Closing TC-loop")
			r.Wg.Done()
			return
		case <-ticker.C:
			value, err := drp.Next()
			numberOfActiveUes := len(sessions)
			if numberOfActiveUes > 1 && shareCapacityResources {
				value /= float64(numberOfActiveUes)
			}
			if err != nil {
				if _, ok := err.(*errortypes.IterableStopError); ok {
					r.Application_has_finished <- "DataRatePattern has finished"
					r.Wg.Done()
					return
				} else {
					r.ReportWarn(fmt.Errorf("LaunchChangeLoop could retrieve next Value: %w", err))
					r.Wg.Done()
					return
				}
			}
			//change data rate in control file
			if err := s.tc.ChangeMultiTo(value); err != nil {
				r.ReportFatal(fmt.Errorf("LaunchChangeLoop could not change value: %w", err))
				r.Wg.Done()
				return
			}
		}
	}
}

func (s *DrpPlayer) playWarmupTime() {
	select {
	case <-time.After(time.Millisecond * time.Duration(s.session.ChildDRP.WarmupTimeMs)):
		return
	case <-s.r.Exit_now_signal:
		return
	}
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
		exit_handler := make(chan os.Signal, 1)
		signal.Notify(exit_handler, syscall.SIGINT, syscall.SIGPIPE, syscall.SIGQUIT)

		//Exit listener

		select {
		case sig := <-exit_handler:
			INFO.Printf("Received Signal: %d", sig)
			s.ExitNoWait()
			INFO.Printf("Resources resetted\n")
		case msg := <-s.r.Application_has_finished:
			INFO.Println(msg)
			DEBUG.Println("Exiting: Application_has_finished")
			s.ExitNoWait()
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
			return

		}

		INFO.Println("Waiting for palyer")
		s.Wait()
		INFO.Println("Finished Waiting for palyer")

	}()
	if s.multisession.DrpMode {
		INFO.Printf("play data rate pattern %s on dev %s with %d samples/s in loop mode %t\n", s.session.ChildDRP.GetName(), s.session.Dev, s.session.ChildDRP.Freq, s.session.ChildDRP.IsLooping())
	}
	// create queues
	if err := s.initTC(); err != nil {
		return fmt.Errorf("initTC returned %w", err)
	}

	s.playWarmupTime()

	db, err := persistence.GetPersistence()
	if err != nil {
		WARN.Println(err)
		return err
	}
	// persist multisession
	err = (*db).Persist(s.multisession)
	if err != nil {
		WARN.Println(err)
		return err
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

	if s.multisession.DrpMode {
		// play drp
		s.r.Wg.Add(1)
		go s.launchMultiChangeLoop(
			time.Duration(1000/s.session.ChildDRP.Freq)*time.Millisecond,
			s.session.ChildDRP,
			s.r,
			s.multisession.ShareCapacityResources,
		)
	}

	//monitor ue load, synchronize queues
	if !s.multisession.SingleQueue && !s.session.Nomeasure {
		s.r.Wg.Add(1)
		go s.monitorTrafficPerUe(db, s.r, s.multisession.ShareCapacityResources)
	}

	return nil
}

func (s *DrpPlayer) launchMeasureSession(i int, db *persistence.Persistence, netflowFilter string, fixedNetflow bool) {
	ueSession := *s.session
	ueSession.Time = uint64(time.Now().UnixMilli())
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

func (s *DrpPlayer) monitorTrafficPerUe(db *persistence.Persistence, r util.RoutineReport, shareCapacityResources bool) {
	sampleDuration := MONITOR_NETFLOW_DURATION_S * time.Second
	monitoringTicker := time.NewTicker(sampleDuration)

	// timer to monitor traffic per UE
	for {
		select {
		case <-monitoringTicker.C:
			// remove non relevant UEs
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
				ueLoadKbits := sumLoadBytes / (125 * uint64(sampleDuration.Seconds()))
				INFO.Printf("ue%v netflow %v %v loadKbits", ueSession.Session.Uenum, ueSession.Session.NetflowFilter, ueLoadKbits)
				if ueLoadKbits < uint64(s.multisession.UeMinloadkbits) {
					ueSession.Wg.Done()
					sessions = append(sessions[:i], sessions[i+1:]...)
					uesChanged = true
					INFO.Printf("ue %v removed", ueSession.Session.Uenum)
				}
			}

			// if traffic in ue0 relevant, launch neu UE
			// get max load in netflow for ue 0
			if uint8(len(sessions)) < s.multisession.UenumTotal {
				ue0Session := sessions[0]
				var maxUe0LoadByte uint64 = 0
				var newNetflow string
				INFO.Printf("check default queue ue0 for new UEs ...")
				for _, aggregated_measure := range ue0Session.AggregateMeasurePerNetflow {
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
					aggregated_measure.SumloadTotalBytes = 0
				}
				// add netflow as ue, if relevant
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
				if !s.multisession.DrpMode {
					// adjust bandwidth per UE
					bandwidthPerUE := float64(s.multisession.Bandwidthkbits)
					if len(sessions) > 1 && shareCapacityResources {
						bandwidthPerUE /= float64(len(sessions))
					}
					//change data rate in control file
					if err := s.tc.ChangeMultiTo(bandwidthPerUE); err != nil {
						r.ReportFatal(fmt.Errorf("monitorTrafficPerUe could not change value: %w", err))
						r.Wg.Done()
						return
					}
				}
			}
		case <-r.Exit_now_signal:
			s.r.Wg.Done()
			INFO.Println("Finishing ip monitorTrafficPerUE")
			return
		}
	}
}

func (s *DrpPlayer) Exit_clean() {
	if err := s.tc.Close(); err != nil {
		WARN.Printf("Exit: error closing TrafficControl: %+v", err)
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
	close(s.r.Exit_now_signal)
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
	var initialDatarateKbits float64
	if s.multisession.DrpMode {
		initialDatarateKbits = float64(s.session.ChildDRP.Peek() * 1.33)
	} else {
		initialDatarateKbits = float64(s.multisession.Bandwidthkbits)
		if !s.multisession.SingleQueue {
			initNumberOfUes := len(s.multisession.FixedNetflows) + 1
			if s.multisession.ShareCapacityResources && initNumberOfUes > 1 {
				initialDatarateKbits /= float64(initNumberOfUes)
			}
		}
	}
	settings := trafficcontrol.TrafficControlStartParams{
		Datarate:     uint32(initialDatarateKbits),
		QueueSize:    int(s.session.Queuesizepackets),
		AddonLatency: int(s.session.ExtralatencyMs),
		Markfree:     int(s.session.Markfree),
		Markfull:     int(s.session.Markfull),
		Qosmode:      s.session.Qosmode,
		Uenum:        s.multisession.UenumTotal + 1,
	}
	DEBUG.Printf("Init Tc: %+v", settings)
	err := s.tc.InitMultijens(settings,
		trafficcontrol.NftStartParams{
			L4sPremarking: s.session.L4sEnablePreMarking,
			SignalStart:   s.session.SignalDrpStart,
			Uenum:         s.multisession.UenumTotal + 1,
			SingleQueue:   s.multisession.SingleQueue,
		},
		s.multisession.FixedNetflows,
	)
	return err
}
