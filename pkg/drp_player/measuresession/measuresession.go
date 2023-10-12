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

package measuresession

/*
 #include "stdio.h"
 #include "stdlib.h"
 #include "poll.h"
 #include <time.h>
 static unsigned long long get_nsecs(void)
 {
	 struct timespec ts;
	 clock_gettime(CLOCK_MONOTONIC, &ts);
	 return (unsigned long long)ts.tv_sec * 1000000000UL + ts.tv_nsec;
 }
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/telekom/aml-jens/internal/logging"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
	"github.com/telekom/aml-jens/pkg/drp_player/trafficcontrol"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

type DB_measure_packet = datatypes.DB_measure_packet
type DB_network_flow = datatypes.DB_network_flow
type DB_measure_queue = datatypes.DB_measure_queue

type AggregateMeasure struct {
	sampleCount       uint32
	sumSojournTimeMs  uint32
	sumloadBytes      uint64
	sumCapacityKbits  int64
	sumEcnNCE         uint32
	sumDropped        uint32
	Net_flow          *datatypes.DB_network_flow
	t_start           uint64
	t_end             uint64
	SumloadTotalBytes uint64
}

var currentCapacityKbits uint64

func (s *AggregateMeasure) AsRule() string {
	return fmt.Sprintf("ip saddr %s %s sport %d ip daddr %s %s dport %d",
		s.Net_flow.Source_ip,
		s.Net_flow.TransportProtocoll,
		s.Net_flow.Source_port,
		s.Net_flow.Destination_ip,
		s.Net_flow.TransportProtocoll,
		s.Net_flow.Destination_port)
}
func (s *AggregateMeasure) toDB_measure_packet(time uint64) DB_measure_packet {
	var sampleCapacityKbits uint64
	sample_duration := util.MaxInt(SAMPLE_DURATION_MS, int(s.t_end-s.t_start))
	loadKbits := ((s.sumloadBytes * 8) / 1000) * (1000 / uint64(sample_duration))
	if s.sumCapacityKbits == -1 {
		sampleCapacityKbits = loadKbits
	} else {
		sampleCapacityKbits = uint64(s.sumCapacityKbits) / uint64(s.sampleCount)
	}
	return DB_measure_packet{
		Time:                time,
		PacketSojournTimeMs: s.sumSojournTimeMs / s.sampleCount,
		LoadKbits:           uint32(loadKbits),
		Ecn:                 uint32((float32(s.sumEcnNCE) / float32(s.sampleCount)) * 100),
		Dropped:             s.sumDropped,
		Fk_flow_id:          s.Net_flow.Flow_id,
		Capacitykbits:       uint32(sampleCapacityKbits),
		Net_flow_string:     s.Net_flow.MeasureIdStr(),
		Net_flow_prio:       s.Net_flow.Prio,
	}
}
func (s *AggregateMeasure) reset() {
	s.sumloadBytes = 0
	s.sumDropped = 0
	s.sumEcnNCE = 0
	s.sumCapacityKbits = 0
	s.sumSojournTimeMs = 0
	s.sampleCount = 0
	s.t_start = 0
}

func NewAggregateMeasure(flow *datatypes.DB_network_flow) *AggregateMeasure {
	return &AggregateMeasure{
		sumloadBytes:      0,
		sumDropped:        0,
		sumEcnNCE:         0,
		sumSojournTimeMs:  0,
		sampleCount:       0,
		Net_flow:          flow,
		t_start:           0,
		t_end:             0,
		SumloadTotalBytes: 0,
	}
}

func (s *AggregateMeasure) add(pm *PacketMeasure) {
	if s.t_start == 0 {
		// start of the sample intervall is end of the previous one
		s.t_start = s.t_end
		if s.t_end == 0 {
			s.t_start = pm.timestampMs
		}
	}
	s.t_end = pm.timestampMs
	if pm.drop {
		s.sumDropped++
		//DEBUG.Printf("D: %+v", pm)
		return
	}
	// aggregate sample values
	s.sumSojournTimeMs += pm.sojournTimeMs
	s.sumloadBytes += uint64(pm.packetSizeByte)
	s.SumloadTotalBytes += uint64(pm.packetSizeByte)
	//Fix for setting capacity to maximum
	s.sumCapacityKbits += int64(pm.currentCapacityKbits)

	if pm.ecnOut == 3 {
		s.sumEcnNCE++
	}
	s.sampleCount++
}

type DbPacketMeasure struct {
	timestampMs   uint64
	sojournTimeMs uint32
	loadKbits     uint32
	capacityKbits uint32
	ecnCePercent  uint32
	dropped       uint32
	netFlow       string
}

type DbQueueMeasure struct {
	timestampMs            uint64
	numberOfPacketsInQueue uint16
	memUsageBytes          uint32
}

const SAMPLE_DURATION_MS = 10

type MeasureSession struct {
	Session                    *datatypes.DB_session
	tc                         *trafficcontrol.TrafficControl
	chan_to_aggregation        chan PacketMeasure
	chan_to_persistence        chan interface{}
	time_diff                  uint64
	Wg                         *sync.WaitGroup
	should_end                 bool
	Persistor                  *MeasureSessionPersistor
	queueFilename              string
	AggregateMeasurePerNetflow map[string]*AggregateMeasure
	FixedNetflow               bool
}

func NewMeasureSession(session *datatypes.DB_session, tc *trafficcontrol.TrafficControl, queueFilename string, fixedNetflow bool) *MeasureSession {
	monotonicMs := uint64(C.get_nsecs()) / 1e6
	systemMs := uint64(time.Now().UnixMilli())
	var wg sync.WaitGroup
	p, err := NewMeasureSessionPersistor(session)
	if err != nil {
		WARN.Printf("Could not create persisitor: %v", err)
	}

	return &MeasureSession{
		Session:                    session,
		tc:                         tc,
		chan_to_aggregation:        make(chan PacketMeasure, 10000),
		chan_to_persistence:        make(chan interface{}, 10000),
		time_diff:                  systemMs - monotonicMs,
		Wg:                         &wg,
		should_end:                 false,
		Persistor:                  p,
		queueFilename:              queueFilename,
		AggregateMeasurePerNetflow: make(map[string]*AggregateMeasure),
		FixedNetflow:               fixedNetflow,
	}

}
func (m *MeasureSession) Stop() {
	m.should_end = true
	DEBUG.Printf("should_end=true (%p)\n", m)
	m.Persistor.Stop()
}
func (m *MeasureSession) Start(r util.RoutineReport) {

	DEBUG.Printf("[rWG+]MeasureSession.Start (%p)\n", &m)
	r.Wg.Add(1)

	// open memory file stream
	INFO.Printf("start measure Session: %s\n", m.Session.Name)

	// compute offset of monotonic and system clock

	// start aggregate and persist threads with communication channels
	DEBUG.Printf("[mWG+] Persistor (%p)\n", &m)
	m.Wg.Add(1)
	defer func() {
		r.Wg.Done()
		DEBUG.Printf("[rWG-]MeasureSession.Start (%p)\n", &m)
	}()
	go m.Persistor.Run(m.chan_to_persistence, func(err error, level util.ErrorLevel) {
		r.Send_error_c <- struct {
			Err   error
			Level util.ErrorLevel
		}{
			Err:   err,
			Level: level,
		}
	}, func() {
		DEBUG.Printf("[mWG-] Closing Persistor (%p).(%p)\n", &m, &m.Persistor)
		m.Wg.Done()
	})

	go m.poll(r, m.queueFilename)

	go m.aggregateMeasures(r)
	DEBUG.Printf("[mWG?] Waiting (%p)\n", &m)
	m.Wg.Wait()
	DEBUG.Printf("[mWG0] Closed measure_session %s (%v)\n", m.Session.Name, &m)
}

// Represents the polling loop.
// Will close if membervariable m.shouldEnd becomes true
// Will foreward this signal by closing chan_to_aggregation
func (m *MeasureSession) poll(r util.RoutineReport, measureFileName string) {
	DEBUG.Println("[mWG+]poll")
	m.Wg.Add(1)
	//Buffer in which the contets of MM_FILE will be written
	recordArray := make(RecordArray, RECORD_SIZE)
	var file, err = os.Open(measureFileName)
	if err != nil {
		r.ReportFatal(fmt.Errorf("measuresession.poll: %w", err))
	}
	/* Clear recordArray due to records in WarmupTime*/
	if m.Session.ChildDRP.WarmupTimeMs > 0 {
		dummy := make([]byte, 0xfffff)
		file.Read(dummy)
		dummy = nil
	}
	defer func() {
		INFO.Println("Closed: Poll")
		if file != nil {
			file.Close()
		}
		m.Wg.Done()
		DEBUG.Println("[mWG-]poll")
		//Forward closing to aggregation
		close(m.chan_to_aggregation)
	}()
	var fdint C.int = C.int(uint(file.Fd()))
	pfd := C.struct_pollfd{fdint, C.POLLIN, 0}
	for !m.should_end {
		// poll fd
		rc := C.poll(&pfd, 1, 1000)
		if rc <= 0 {
			continue
		}
		// read one record of either packet or queue type
		bytesRead, err := file.Read(recordArray)
		if err == io.EOF {
			r.ReportInfo(fmt.Errorf("EOF while reading recordArray"))
		}

		if bytesRead != 64 {
			r.ReportInfo(fmt.Errorf("bytesRead != 64 while reading recordArray %+v", m.should_end))
			continue
		}
		timestampMs := uint64(binary.LittleEndian.Uint64(recordArray[0:8])) / 1e6
		switch recordArray.type_id() {
		case RECORD_TYPE_P: // PacketMeasure MP
			packetMeasure, err := recordArray.AsPacketMeasure(m.Session.Session_id)
			if err != nil {
				r.ReportWarn(fmt.Errorf("could not parse packetMeasure: %w", err))
			}
			if packetMeasure != nil {
				packetMeasure.currentCapacityKbits = currentCapacityKbits
				m.chan_to_aggregation <- *packetMeasure

			} else {
				//DEBUG.Printf("non ip packet ignored\n")
			}
		case RECORD_TYPE_Q: // QueueMeasure MQ
			numberOfPacketsInQueue := uint16(binary.LittleEndian.Uint16(recordArray[10:12]))
			memUsageBytes := uint32(binary.LittleEndian.Uint32(recordArray[12:16]))
			currentCapacityKbits = uint64(binary.LittleEndian.Uint64(recordArray[16:24])) / 1000
			currentEpochMs := timestampMs + m.time_diff
			queueMeasure := DB_measure_queue{
				Time:              currentEpochMs,
				Memoryusagebytes:  memUsageBytes,
				PacketsInQueue:    numberOfPacketsInQueue,
				CapacityKbits:     currentCapacityKbits,
				Fk_session_tag_id: m.Session.Session_id,
			}
			if m.should_end {
				return

			}
			m.chan_to_persistence <- queueMeasure
		default: //Error
			r.ReportWarn(fmt.Errorf("could not parse record_array (type): %+v", recordArray))
		}
	}
}
func (m MeasureSession) aggregateMeasures(r util.RoutineReport) {
	DEBUG.Printf("[mWG+]aggregateMeasures (%p)\n", &m)
	m.Wg.Add(1)
	sampleDuration := SAMPLE_DURATION_MS * time.Millisecond
	ticker := time.NewTicker(sampleDuration)
	defer func() {
		DEBUG.Printf("[mWG-]aggregateMeasures (%p)\n", &m)
		close(m.chan_to_persistence)
		m.Wg.Done()
	}()
	doExit := false
	p, err := persistence.GetPersistence()
	if err != nil {
		r.ReportFatal(fmt.Errorf("aggregateMeasure: %w", err))
		return
	}
	for range ticker.C {
		readMessages := true
		var packetStartTimeMs uint64 = 0
		message, is_open := <-m.chan_to_aggregation
		if !is_open {
			return
		} else {
			packetStartTimeMs = message.timestampMs
		}

		for readMessages {
			select {
			case message, is_open = <-m.chan_to_aggregation: // Aggregate Measure

				if !is_open {
					readMessages = false
					doExit = true
					break
				}
				diffMs := int64(message.timestampMs - packetStartTimeMs)
				if diffMs >= sampleDuration.Milliseconds() {
					readMessages = false
				}
				measure, keyExists := m.AggregateMeasurePerNetflow[message.net_flow.MeasureIdStr()]
				if !keyExists {
					measure = NewAggregateMeasure(message.net_flow)
					m.AggregateMeasurePerNetflow[message.net_flow.MeasureIdStr()] = measure

					if err := (*p).Persist(message.net_flow); err != nil {
						r.ReportFatal(fmt.Errorf("aggregateMeasure: %w", err))
						return
					}
				}

				measure.add(&message)
			default:
				readMessages = false
			}
		}
		if doExit {
			DEBUG.Println("Returning from aggregation")
			return
		}
		for _, aggregated_measure := range m.AggregateMeasurePerNetflow {
			if aggregated_measure.sampleCount == 0 {
				continue
			}
			// send to persist measure sample
			currentEpochMs := message.timestampMs + m.time_diff
			sample := aggregated_measure.toDB_measure_packet(currentEpochMs)
			sample.Uenum = m.Session.Uenum
			if m.Session.ParentBenchmark.CsvOuptut {
				if sample.Capacitykbits == 0 {
					//this sometimes happens
					//Everything in the sample is zero but the time
					if sample.Dropped != 0 || sample.LoadKbits != 0 || sample.Ecn != 0 {
						INFO.Printf("Capacity is 0 but not everything: %+v", sample)
						continue
					}
					DEBUG.Printf("Not Persisting: %+v", sample)
					continue
				}
			}
			m.chan_to_persistence <- sample
			// reset temporary sum fields
			aggregated_measure.reset()
		}
	}
}
