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
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
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

type PacketMeasure struct {
	//internal use, as unique ID
	timestampMs    uint64
	sojournTimeMs  uint32
	ecnIn          uint8
	ecnOut         uint8
	ecnValid       bool
	slow           bool
	mark           bool
	drop           bool
	ipVersion      uint8
	packetSizeByte uint32
	net_flow       *datatypes.DB_network_flow
}

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

const MM_FILE = "/sys/kernel/debug/sch_janz/0001:0"

const RECORD_SIZE = 64
const QUEUE_DATA = 6
const PACKET_DATA = 7
const TC_JENS_RELAY_ECN_VALID = 1 << 2
const TC_JENS_RELAY_SOJOURN_SLOW = 1 << 5
const TC_JENS_RELAY_SOJOURN_MARK = 1 << 6
const TC_JENS_RELAY_SOJOURN_DROP = 1 << 7

const SAMPLE_DURATION_MS = 10

var bool2int = map[bool]int8{false: 0, true: 1}

type exitStruct struct{}

var WaitGroup sync.WaitGroup

// Start the measuresession
//
// # Uses util.RoutineReport
//
// Blocking - spawns 3 (three)  more goroutines, adds them to wg
func Start(session *datatypes.DB_session, tc *trafficcontrol.TrafficControl, r util.RoutineReport) {
	// open memory file stream
	recordArray := make([]byte, RECORD_SIZE)
	INFO.Println("start measure session")
	var file, err = os.Open(MM_FILE)
	if err != nil {
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
		r.Send_error_c <- err
		r.Wg.Done()
		return
	}
	r.Wg.Add(2)
	go aggregateMeasures(session, chan_pers_measure, chan_pers_sample, diffTimeMs, tc, r)
	go persistor.Run(r, chan_pers_sample)
	var fdint C.int = C.int(uint(file.Fd()))
	pfd := C.struct_pollfd{fdint, C.POLLIN, 0}
	r.Wg.Add(1)
	go func() {
		for {
			select {
			case <-r.On_extern_exit_c:
				DEBUG.Println("Closing measuresession")
				persistor.Close()
				r.Wg.Done()
				return
			case <-time.After(10 * time.Millisecond):
				rc := C.poll(&pfd, 1, 9)
				if rc > 0 {
					bytesRead, err := file.Read(recordArray)
					if err == io.EOF {
						INFO.Println("")
						INFO.Println("EOF")
						INFO.Println("")
					}
					chan_poll_result <- bytesRead
				}
			}
		}
	}()
	for {
		select {
		case <-r.On_extern_exit_c:
			DEBUG.Println("Closing measuresession - recordarray parser")
			persistor.Close()
			r.Wg.Done()
			return
		case bytesRead := <-chan_poll_result:
			// read one record of either packet or queue type
			if bytesRead != 64 {
				continue
			}
			timestampMs := uint64(binary.LittleEndian.Uint64(recordArray[0:8])) / 1e6
			packetType := recordArray[8]

			if packetType == PACKET_DATA {
				sojournTimeMs := uint32(binary.LittleEndian.Uint32(recordArray[12:16])) / 1e3
				ecnIn := recordArray[9] & 3
				ecnOut := (recordArray[9] & 24) >> 3
				ecnValid := (recordArray[9] & TC_JENS_RELAY_ECN_VALID) != 0
				slow := (recordArray[9] & TC_JENS_RELAY_SOJOURN_SLOW) != 0
				mark := (recordArray[9] & TC_JENS_RELAY_SOJOURN_MARK) != 0
				drop := (recordArray[9] & TC_JENS_RELAY_SOJOURN_DROP) != 0
				ipVersion := recordArray[52]
				srcIp := fmt.Sprintf("%d.%d.%d.%d", uint8(recordArray[28]), uint8(recordArray[29]), uint8(recordArray[30]), uint8(recordArray[31]))
				dstIp := fmt.Sprintf("%d.%d.%d.%d", uint8(recordArray[44]), uint8(recordArray[45]), uint8(recordArray[46]), uint8(recordArray[47]))
				packetSize := uint32(binary.LittleEndian.Uint32(recordArray[48:52]))
				nextHdr := recordArray[53]
				var srcPort uint16 = 0
				var dstPort uint16 = 0
				if nextHdr == 6 || nextHdr == 17 {
					srcPort = uint16(binary.LittleEndian.Uint16(recordArray[54:56]))
					dstPort = uint16(binary.LittleEndian.Uint16(recordArray[56:58]))
				}
				if srcIp != "0.0.0.0" && dstIp != "0.0.0.0" {
					flow := datatypes.DB_network_flow{
						Source_ip:        srcIp,
						Source_port:      srcPort,
						Destination_ip:   dstIp,
						Destination_port: dstPort,
						Session_id:       session.Session_id,
					}
					packetMeasure := PacketMeasure{
						timestampMs:    timestampMs,
						sojournTimeMs:  sojournTimeMs,
						ecnIn:          ecnIn,
						ecnOut:         ecnOut,
						ecnValid:       ecnValid,
						slow:           slow,
						mark:           mark,
						drop:           drop,
						ipVersion:      ipVersion,
						packetSizeByte: packetSize,
						net_flow:       &flow,
					}
					chan_pers_measure <- packetMeasure
				} else {
					INFO.Printf("non ip packet ignored\n")
				}
			} else if packetType == QUEUE_DATA {
				// to do: persistance
				numberOfPacketsInQueue := uint16(binary.LittleEndian.Uint16(recordArray[10:12]))
				memUsageBytes := uint32(binary.LittleEndian.Uint32(recordArray[12:16]))
				currentEpochMs := timestampMs + diffTimeMs
				queueMeasure := DB_measure_queue{
					Time:              currentEpochMs,
					Memoryusagebytes:  memUsageBytes,
					PacketsInQueue:    numberOfPacketsInQueue,
					Fk_session_tag_id: session.Session_id,
				}

				chan_pers_sample <- queueMeasure
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
	// stdout heading
	if session.ParentBenchmark.PrintToStdOut {
		fmt.Println(strings.Join(assets.CONST_HEADING, " "))
	}
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
					FATAL.Exit(err)
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
	defer func() {
		DEBUG.Println("Defer persist_samples <- exitStruct{}  PRE")
		persist_samples <- exitStruct{}
		DEBUG.Println("Defer persist_samples <- exitStruct{}  POST")
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
				if session.ParentBenchmark.PrintToStdOut {
					if err := sample.PrintLine(aggregated_measure.net_flow.MeasureIdStr()); err != nil {
						WARN.Println("Could not write Measurement")
						r.Send_error_c <- err
						r.Wg.Done()
						return false
					}
				}
				return true
			})
			mapMeasures.Range(func(key, value any) bool {
				mapMeasures.Delete(key)
				return true
			})
		}
	}
}

func persistMeasures(session *datatypes.DB_session, samples chan interface{}, r util.RoutineReport) {
	tickerPersist := time.NewTicker(1 * time.Second)

	var db, err = persistence.GetPersistence()
	if err != nil {
		r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
	}
	//var db sql.DB
	// create session tag
	//check if tag already present

	//prepare csv output
	var csvFile *os.File
	var csvWriter *csv.Writer
	var csvQueueFile *os.File
	var csvQueueWriter *csv.Writer
	if session.ParentBenchmark.CsvOuptut {
		if err = os.Mkdir(session.Name, os.ModeDir); err != nil {
			if os.IsExist(err) {
				INFO.Printf("directory %s exists,", session.Name)
				session.Name = session.Name + time.Now().Format("_15:04:05")
				INFO.Printf("storing csv measures in directory %s \n", session.Name)
				err = os.Mkdir(session.Name, os.ModeDir)
				if err != nil {
					r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
				}
			}
		}

		csvFile, err = os.Create(filepath.Join(session.Name, filepath.Base("measure_packet.csv")))
		if err != nil {
			r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
		}
		defer csvFile.Close()
		csvWriter = csv.NewWriter(csvFile)
		heading := assets.CONST_HEADING
		if err := csvWriter.Write(heading); err != nil {
			r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
		}

		csvQueueFile, err = os.Create(filepath.Join(session.Name, filepath.Base("measure_queue.csv")))
		if err != nil {
			r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
		}
		defer csvQueueFile.Close()
		csvQueueWriter = csv.NewWriter(csvQueueFile)
		heading = []string{"timestampMs", "memUsageBytes", "packetsinqueue"}
		if err := csvQueueWriter.Write(heading); err != nil {
			r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
		}
	}
	for {
		select {
		case <-r.On_extern_exit_c:
			r.Wg.Done()
			return
		case <-tickerPersist.C: //start := time.Now()

			readSamples := true
			var samplePacketCount uint32 = 0
			var sampleQueueCount uint32 = 0
			var avgLoadKbits uint32 = 0
			for readSamples {
				select {
				case sampleInterface := <-samples:
					switch sample := sampleInterface.(type) {
					// write measure to db
					case DB_measure_packet:
						if err := (*db).Persist(sample); err != nil {
							FATAL.Exit(err)
						}

						if session.ParentBenchmark.CsvOuptut {
							if err := csvWriter.Write(sample.CsvRecord()); err != nil {
								FATAL.Exitln("error writing record to csv file", err)
							}
						}

						avgLoadKbits += sample.LoadKbits
						samplePacketCount++
					case DB_measure_queue:
						if err := (*db).Persist(&sample); err != nil {
							FATAL.Exitf("---Persist(%+v) -> %s", sample, err)
						}

						if session.ParentBenchmark.CsvOuptut {
							if err := csvQueueWriter.Write(sample.CsvRecord()); err != nil {
								FATAL.Exitln("error writing record to queue csv file", err)
							}
						}

						sampleQueueCount++
					case exitStruct:
						INFO.Println("Exiting persistMeasures: due to struct")
						r.Wg.Done()
						return
					default:
						FATAL.Println("Unexpected Input in persistMeasures")
					}
				default:
					readSamples = false
				}
			}
			(*db).Commit()
			//to csv
			if session.ParentBenchmark.CsvOuptut {
				csvWriter.Flush()
				csvQueueWriter.Flush()
			}

			//durationMs := time.Now().UnixMicro() - start.UnixMicro()

			if samplePacketCount > 0 {
				//avgTimePersistPerMeasure := durationMs / int64(samplePacketCount)
				avgLoadKbits /= samplePacketCount
				//INFO.Printf("samples persisted: type packet %d type queue %d avg time %d us avg load %d\n", samplePacketCount, sampleQueueCount, avgTimePersistPerMeasure, avgLoadKbits)
			}
		}
	}
}
