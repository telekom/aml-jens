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
	sampleCount      uint32
	sumSojournTimeMs uint32
	sumloadBytes     uint32
	sumCapacityKbits float64
	sumEcnNCE        uint32
	sumDropped       uint32
	net_flow         *datatypes.DB_network_flow
}

func (s *AggregateMeasure) toDB_measure_packet(time uint64) DB_measure_packet {
	var sampleCapacityKbits uint32
	loadKbits := ((s.sumloadBytes * 8) / 1000) * (1000 / SAMPLE_DURATION_MS)
	if s.sumCapacityKbits == -1 {
		sampleCapacityKbits = loadKbits
	} else {
		sampleCapacityKbits = uint32(s.sumCapacityKbits) / s.sampleCount
	}
	return DB_measure_packet{
		Time:                time,
		PacketSojournTimeMs: s.sumSojournTimeMs / s.sampleCount,
		LoadKbits:           loadKbits,
		Ecn:                 uint32((float32(s.sumEcnNCE) / float32(s.sampleCount)) * 100),
		Dropped:             s.sumDropped,
		Fk_flow_id:          s.net_flow.Flow_id,
		Capacitykbits:       sampleCapacityKbits,
		Net_flow_string:     s.net_flow.MeasureIdStr(),
	}
}

func NewAggregateMeasure(flow *datatypes.DB_network_flow) *AggregateMeasure {
	return &AggregateMeasure{sumloadBytes: 0, sumDropped: 0, sumEcnNCE: 0, sumSojournTimeMs: 0, sampleCount: 0, net_flow: flow}
}

func (s *AggregateMeasure) add(pm *PacketMeasure, capacity float64) {

	// aggregate sample values
	s.sumSojournTimeMs += pm.sojournTimeMs
	s.sumloadBytes += pm.packetSizeByte
	//Fix for setting capacity to maximum
	if capacity == 4294967295 {
		//if dummy capacity.
		s.sumCapacityKbits = -1
		//s.sumCapacityKbits += float64(pm.packetSizeByte)

	} else {
		s.sumCapacityKbits += capacity
	}

	if pm.ecnOut == 3 {
		s.sumEcnNCE++
	}
	s.sumDropped += uint32(bool2int[pm.drop])
	s.sampleCount++
}

const MM_FILE = "/sys/kernel/debug/sch_janz/0001:0"

const SAMPLE_DURATION_MS = 10

var bool2int = map[bool]int8{false: 0, true: 1}

// Start the measuresession
//
// # Uses util.RoutineReport
//
// Blocking - spawns 3 (three)  more goroutines, adds them to wg
func Start(session *datatypes.DB_session, tc *trafficcontrol.TrafficControl, r util.RoutineReport) {
	// open memory file stream
	recordArray := make(RecordArray, RECORD_SIZE)
	INFO.Println("start measure session")
	var file, err = os.Open(MM_FILE)
	if err != nil {
		FATAL.Println("Could not open MM_FILE")
		r.Send_error_c <- err
		r.Wg.Done()
		return
	}
	defer file.Close()

	recordCount := 0

	// compute offset of monotonic and system clock
	monotonicMs := uint64(C.get_nsecs()) / 1e6
	systemMs := uint64(time.Now().UnixMilli())
	diffTimeMs := systemMs - monotonicMs

	// start aggregate and persist threads with communication channels
	chan_pers_measure := make(chan PacketMeasure, 10000)
	chan_pers_sample := make(chan interface{}, 10000)
	chan_poll_result := make(chan int)
	persistor, err := NewMeasureSessionPersistor(session)
	if err != nil {
		FATAL.Println("Could not create Persistor")
		r.Send_error_c <- err
		r.Wg.Done()
		return
	}
	r.Wg.Add(1)
	go persistor.Run(r, chan_pers_sample)
	r.Wg.Add(1)
	go aggregateMeasures(session, chan_pers_measure, chan_pers_sample, diffTimeMs, tc, r)

	var fdint C.int = C.int(uint(file.Fd()))
	pfd := C.struct_pollfd{fdint, C.POLLIN, 0}
	//Helper function
	done := func() {
		persistor.Close()
		r.Wg.Done()
	}
	r.Wg.Add(1)
	go func() {
		for {
			select {
			case <-r.On_extern_exit_c:
				DEBUG.Println("Closing measuresession")
				r.Wg.Done()
				return
			case <-time.After(10 * time.Millisecond):
				rc := C.poll(&pfd, 1, 9)
				if rc > 0 {
					bytesRead, err := file.Read(recordArray)
					if err == io.EOF {
						INFO.Println("EOF")
					}
					select {
					case chan_poll_result <- bytesRead:
						//Good
					default:
						// chan_poll_result does not get polled
						DEBUG.Println("Throwing away poll event")
					}

				}
			}
		}
	}()

	for {
		select {

		case <-r.On_extern_exit_c:
			DEBUG.Println("Closing measuresession - recordarray parser")
			done()
			return
		case bytesRead := <-chan_poll_result:
			// read one record of either packet or queue type
			if bytesRead != RECORD_SIZE {
				continue
			}
			switch recordArray.type_id() {
			case RECORD_TYPE_Q:
				queueMeasure, err := recordArray.AsDB_measure_queue(diffTimeMs, session.Session_id)
				if err != nil {
					WARN.Println("Could not format record as MQ")
					r.Send_error_c <- err
					done()
					return
				}
				chan_pers_sample <- *queueMeasure
			case RECORD_TYPE_P:
				packetMeasure, err := recordArray.AsPacketMeasure(session.Session_id)
				if err != nil {
					WARN.Println("Could not format record as MP")
					r.Send_error_c <- err
					done()
					return
				}
				if packetMeasure == nil {
					INFO.Printf("non ip packet ignored\n")
				} else {
					chan_pers_measure <- *packetMeasure
				}
			default:
				WARN.Printf("Cant Parse recordarray %v: unknown type", recordArray)
			}

			recordCount++
		}
	}
}

// Start the measuresession
//
// # Uses util.RoutineReport
//
// Blocking - spawns 1 (one)  more goroutine, adds them to wg
func aggregateMeasures(session *datatypes.DB_session, messages chan PacketMeasure, persist_samples chan interface{}, diffTimeMs uint64, tc *trafficcontrol.TrafficControl, r util.RoutineReport) {
	sampleDuration := SAMPLE_DURATION_MS * time.Millisecond
	ticker := time.NewTicker(sampleDuration)
	//Use of sync.Map instead of map:
	//Used as a cache
	mapMeasures := sync.Map{}
	r.Wg.Add(1)
	go func() {
		p, err := persistence.GetPersistence()
		if err != nil {
			r.Send_error_c <- fmt.Errorf("aggregateMeasure: %w", err)
			r.Wg.Done()
			return
		}
		for {
			select {
			case <-r.On_extern_exit_c:
				DEBUG.Println("Closing aggregateMeasures - writer")
				r.Wg.Done()
				return
			case single_packet_measure := <-messages:
				if err := (*p).Persist(single_packet_measure.net_flow); err != nil {
					r.Send_error_c <- err
				}
				measure, _ := mapMeasures.LoadOrStore(single_packet_measure.net_flow.MeasureIdStr(), NewAggregateMeasure(single_packet_measure.net_flow))
				m, ok := measure.(*AggregateMeasure)
				if !ok {
					//This should NEVER happen
					r.Send_error_c <- fmt.Errorf("stored %+v in mapMeasures", measure)
					r.Wg.Done()
					return
				}
				m.add(&single_packet_measure, tc.CurrentRate())
			}
		}
	}()
	for {
		select {
		case <-r.On_extern_exit_c:
			DEBUG.Println("Closing aggregateMeasures - reader")
			r.Wg.Done()
			return
		case <-ticker.C:
			mapMeasures.Range(func(key any, value any) bool {
				aggregated_measure, ok := value.(*AggregateMeasure)
				if !ok {
					WARN.Println("Can NOT iterate Measures, got invalid value.")
					return false
				}
				if aggregated_measure.sampleCount == 0 {
					return true
				}
				currentEpochMs := time.Now().UnixNano()/(int64(time.Millisecond)/int64(time.Nanosecond)) - 5

				sample := aggregated_measure.toDB_measure_packet(uint64(currentEpochMs))
				if sample.Capacitykbits == 0 {
					//this sometimes happens
					//Everything in the sample is zero but the time
					if sample.Dropped != 0 || sample.LoadKbits != 0 || sample.Ecn != 0 {
						DEBUG.Printf("Capacity is 0 but not everything: %+v", sample)
						return true
					}
					INFO.Printf("Not Persisting: %+v", sample)
					return true
				}
				persist_samples <- sample
				return true
			})
			mapMeasures.Range(func(key, value any) bool {
				mapMeasures.Delete(key)
				return true
			})

		}
	}
}
