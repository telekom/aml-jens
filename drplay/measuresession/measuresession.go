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
	"jens/drcommon/assets"
	"jens/drcommon/logging"
	"jens/drcommon/persistence"
	"jens/drcommon/persistence/datatypes"
	"jens/drplay/trafficcontrol"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var DEBUG, INFO, FATAL = logging.GetLogger()

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

func Start(session *datatypes.DB_session, tc *trafficcontrol.TrafficControl, channel_close chan struct{}) {
	// open memory file stream
	recordArray := make([]byte, RECORD_SIZE)
	INFO.Println("start measure session")
	var file, err = os.Open(MM_FILE)
	if err != nil {
		FATAL.Exit(err)
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

	WaitGroup.Add(2)
	go aggregateMeasures(session, chan_pers_measure, chan_pers_sample, channel_close, diffTimeMs, tc)
	go persistMeasures(session, chan_pers_sample)

	var fdint C.int = C.int(uint(file.Fd()))
	pfd := C.struct_pollfd{fdint, C.POLLIN, 0}
	for {
		// poll fd
		rc := C.poll(&pfd, 1, -1)
		if rc > 0 {
			// read one record of either packet or queue type
			bytesRead, _ := file.Read(recordArray)
			if err == io.EOF {
				INFO.Println("")
				INFO.Println("EOF")
				INFO.Println("")
			}

			if bytesRead == 64 {
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
}

func aggregateMeasures(session *datatypes.DB_session, messages chan PacketMeasure, persist_samples chan interface{}, channel_exit chan struct{}, diffTimeMs uint64, tc *trafficcontrol.TrafficControl) {
	sampleDuration := SAMPLE_DURATION_MS * time.Millisecond
	ticker := time.NewTicker(sampleDuration)

	// stdout heading
	if session.ParentBenchmark.PrintToStdOut {
		fmt.Println(strings.Join(assets.CONST_HEADING, " "))
	}
	doExit := false
	for range ticker.C {
		mapMeasures := make(map[string]*AggregateMeasure)
		readMessages := true
		message := <-messages
		packetStartTimeMs := message.timestampMs
		for readMessages {
			select {
			case <-channel_exit:
				INFO.Println("Closing aggregateMeasures")
				readMessages = false
				doExit = true
				break

			case message = <-messages: // Aggregate Measure
				diffMs := int64(message.timestampMs - packetStartTimeMs)
				if diffMs > sampleDuration.Milliseconds() {
					readMessages = false
				}
				// update measure sample
				p, err := persistence.GetPersistence()
				if err != nil {
					FATAL.Exit(err)
				}
				if err := (*p).Persist(message.net_flow); err != nil {
					FATAL.Exit(err)
				}
				measure, keyExists := mapMeasures[message.net_flow.MeasureIdStr()]
				if !keyExists {
					measure = NewAggregateMeasure(message.net_flow)
					mapMeasures[message.net_flow.MeasureIdStr()] = measure
				}

				measure.add(&message, tc.CurrentRate())

			default:
				readMessages = false
			}
		}
		for _, aggregated_measure := range mapMeasures {
			if aggregated_measure.sampleCount == 0 {
				continue
			}
			// send to persist measure sample
			currentEpochMs := message.timestampMs + diffTimeMs
			sample := aggregated_measure.toDB_measure_packet(currentEpochMs)
			if session.ParentBenchmark.CsvOuptut {
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
			persist_samples <- sample
			if session.ParentBenchmark.PrintToStdOut {
				sample.PrintLine(aggregated_measure.net_flow.MeasureIdStr())
			}
		}
		if doExit {
			persist_samples <- exitStruct{}
			WaitGroup.Done()
			return
		}
	}
}

func persistMeasures(session *datatypes.DB_session, samples chan interface{}) {
	tickerPersist := time.NewTicker(1 * time.Second)

	var db, err = persistence.GetPersistence()
	if err != nil {
		FATAL.Exit(err)
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
					FATAL.Println(err)
				}
			}
		}

		csvFile, err = os.Create(filepath.Join(session.Name, filepath.Base("measure_packet.csv")))
		if err != nil {
			FATAL.Exitln("failed to create csv file", err)
		}
		defer csvFile.Close()
		csvWriter = csv.NewWriter(csvFile)
		heading := assets.CONST_HEADING
		if err := csvWriter.Write(heading); err != nil {
			FATAL.Exitln("error writing heading to csv file", err)
		}

		csvQueueFile, err = os.Create(filepath.Join(session.Name, filepath.Base("measure_queue.csv")))
		if err != nil {
			FATAL.Exitln("failed to create queue csv file", err)
		}
		defer csvQueueFile.Close()
		csvQueueWriter = csv.NewWriter(csvQueueFile)
		heading = []string{"timestampMs", "memUsageBytes", "packetsinqueue"}
		if err := csvQueueWriter.Write(heading); err != nil {
			FATAL.Exitln("error writing heading to queue csv file", err)
		}
	}

	for range tickerPersist.C {
		//start := time.Now()

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
					INFO.Println("Exiting persistMeasures")
					WaitGroup.Done()
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
