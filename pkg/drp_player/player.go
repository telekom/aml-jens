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

type MultiSession_Session struct {
	ms           *measuresession.MeasureSession
	time_created time.Time
	is_permament bool
}

type DrpPlayer struct {
	multisession        *datatypes.DB_multi_session
	session             *datatypes.DB_session
	tc                  *trafficcontrol.TrafficControl
	r                   util.RoutineReport
	is_shutting_down    bool
	close_channel_mutex *sync.Mutex
	ue_s                []*MultiSession_Session
}

func NewDrpPlayer(config *config.DrPlayConfig) *DrpPlayer {
	var wg sync.WaitGroup
	var close_channel_mutex sync.Mutex
	d := &DrpPlayer{
		is_shutting_down: false,
		multisession:     config.A_MultiSession,
		session:          config.A_Session,
		ue_s:             make([]*MultiSession_Session, 0, 16),
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
	DEBUG.Println("[rWG+]launchMultiChangeLoopDone")
	s.r.Wg.Add(1)
	defer func() {
		DEBUG.Println("[rWG-]launchMultiChangeLoopDone")
		r.Wg.Done()
	}()
	INFO.Printf("start playing DataRatePattern @%s", waitTime.String())
	for {
		select {
		case <-r.Exit_now_signal:
			DEBUG.Println("Closing TC-loop")
			return
		case <-ticker.C:
			value, err := drp.Next()
			numberOfActiveUes := len(s.ue_s)
			if numberOfActiveUes > 1 && shareCapacityResources {
				value /= float64(numberOfActiveUes)
			}
			if err != nil {
				if _, ok := err.(*errortypes.IterableStopError); ok {
					r.Application_has_finished <- "DataRatePattern has finished"
					return
				} else {
					r.ReportWarn(fmt.Errorf("LaunchChangeLoop could retrieve next Value: %w", err))
					return
				}
			}
			//change data rate in control file
			if err := s.tc.ChangeMultiTo(value); err != nil {
				r.ReportFatal(fmt.Errorf("LaunchChangeLoop could not change value: %w", err))
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
			DEBUG.Printf("Received Signal: %d", sig)
			s.ExitNoWait()
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
		go s.launchMultiChangeLoop(
			time.Duration(1000/s.session.ChildDRP.Freq)*time.Millisecond,
			s.session.ChildDRP,
			s.r,
			s.multisession.ShareCapacityResources,
		)
	}

	//monitor ue load, synchronize queues
	if !s.multisession.SingleQueue && !s.session.Nomeasure {

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
	new_multi_session := &MultiSession_Session{
		time_created: time.Now(),
		is_permament: fixedNetflow,
	}
	// filename to read queue measures
	thisMeasureFilename := fmt.Sprintf(MEASURE_FILE_MULTIJENS_PRAEFIX+"0%d:0", i)
	new_multi_session.ms = measuresession.NewMeasureSession(&ueSession, s.tc, thisMeasureFilename, fixedNetflow)
	go new_multi_session.ms.Start(s.r)
	s.ue_s = append(s.ue_s, new_multi_session)
}

func (s *DrpPlayer) monitorTrafficPerUe(db *persistence.Persistence, r util.RoutineReport, shareCapacityResources bool) {
	sampleDuration := MONITOR_NETFLOW_DURATION_S * time.Second
	monitoringTicker := time.NewTicker(sampleDuration)
	DEBUG.Println("[rWG+]monitorTrafficPerUe")
	s.r.Wg.Add(1)
	defer func() {
		DEBUG.Println("[rWG-]monitorTrafficPerUe")
		s.r.Wg.Done()
	}()
	// timer to monitor traffic per UE
	for {
		select {
		case <-monitoringTicker.C:
			// remove non relevant UEs
			uesChanged := false
			// Check catchall UE
			// if traffic in ue0 relevant, launch new UE
			// get max load in netflow for ue 0
			if uint8(len(s.ue_s)) < s.multisession.UenumTotal {
				ue0 := s.ue_s[0]
				var maxUe0LoadByte uint64 = 0
				var flow_to_split_name string
				var flow_to_split_rule string
				INFO.Printf("check default queue ue0 for new UEs ...")
				for key, aggregated_measure := range ue0.ms.AggregateMeasurePerNetflow {
					INFO.Printf("Checking: %v", aggregated_measure.Net_flow.MeasureIdStr())
					if aggregated_measure.SumloadTotalBytes > maxUe0LoadByte {
						maxUe0LoadByte = aggregated_measure.SumloadTotalBytes
						flow_to_split_name = key
						flow_to_split_rule = aggregated_measure.AsRule()
					}
					aggregated_measure.SumloadTotalBytes = 0
				}
				// add netflow as ue, if relevant
				avgLoadKbits := maxUe0LoadByte / (125 * uint64(sampleDuration.Seconds()))
				if avgLoadKbits > uint64(s.multisession.UeMinloadkbits) {
					/**
					  * delete(ue0.ms.AggregateMeasurePerNetflow, flow_to_split_name)
					  * It would be good to delete unnecessary Flows.
					  * BUT: This breaks a lot of things.
					  * TODO: Research why
					  *
					**/
					lastSessionIndex := len(s.ue_s)
					INFO.Printf("ADDING ue%v netflow %v %v kbits added", lastSessionIndex, flow_to_split_name, avgLoadKbits)
					s.launchMeasureSession(lastSessionIndex, db, flow_to_split_rule, false)
					uesChanged = true
				}
			}
			ue_copy := make([]*MultiSession_Session, 0, 16)
			for _, ue := range s.ue_s {
				if ue.is_permament || time.Since(ue.time_created) < time.Second*5 {
					ue_copy = append(ue_copy, ue)
					continue
				}

				var sumLoadBytes uint64 = 0
				// A non automatic UE _can_ have more than 1 Aggregate Measure
				for _, aggregated_measure := range ue.ms.AggregateMeasurePerNetflow {
					sumLoadBytes += aggregated_measure.SumloadTotalBytes
					aggregated_measure.SumloadTotalBytes = 0
				}
				//What is this calculation ?
				ueLoadKbits := sumLoadBytes / (125 * uint64(sampleDuration.Seconds()))
				if ueLoadKbits < uint64(s.multisession.UeMinloadkbits/3) {
					INFO.Printf("REMOVING: ue%v netflow %v %v loadKbits", ue.ms.Session.Uenum, ue.ms.Session.NetflowFilter, ueLoadKbits)
					ue.ms.Stop()
					uesChanged = true
					INFO.Printf("ue %v removed", ue.ms.Session.Uenum)
				} else {
					ue_copy = append(ue_copy, ue)
				}
			}
			if uesChanged {
				s.ue_s = ue_copy
				// reinstanciate marking rules
				s.rebuild_marking_rules(shareCapacityResources, r)
			}
		case <-r.Exit_now_signal:
			return
		}
	}
}

func (s *DrpPlayer) rebuild_marking_rules(shareCapacityResources bool, r util.RoutineReport) {
	INFO.Printf("rebuild nft marking rules")
	var netfilter []string
	for _, session := range s.ue_s {
		if session.ms.Session.NetflowFilter != "" {
			netfilter = append(netfilter, session.ms.Session.NetflowFilter)
		}
	}
	trafficcontrol.ResetNFT(assets.NFT_TABLE_UEMARK)
	err := trafficcontrol.CreateRulesMarkUe(netfilter)
	if err != nil {
		FATAL.Println(err)
	}
	if !s.multisession.DrpMode {
		// adjust bandwidth per UE
		bandwidthPerUE := float64(s.multisession.Bandwidthkbits)
		if len(s.ue_s) > 1 && shareCapacityResources {
			bandwidthPerUE /= float64(len(s.ue_s))
		}
		//change data rate in control file
		if err := s.tc.ChangeMultiTo(bandwidthPerUE); err != nil {
			r.ReportFatal(fmt.Errorf("monitorTrafficPerUe could not change value: %w", err))
			s.r.Wg.Done()
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
	DEBUG.Println("[rWG?] waiting")
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
	for _, session := range s.ue_s {
		session.ms.Stop()
	}
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
	s.tc.Reset()
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
