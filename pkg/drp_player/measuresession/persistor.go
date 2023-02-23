package measuresession

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
)

type csvPersistorData struct {
	PacketFile   *os.File
	PacketWriter *csv.Writer
	QueueFile    *os.File
	QueueWriter  *csv.Writer
}

type MeasureSessionPersistor struct {
	session           *datatypes.DB_session
	persist_frequency time.Duration
	db                *persistence.Persistence
	csv               *csvPersistorData
}

func NewMeasureSessionPersistor(
	session *datatypes.DB_session,
) (*MeasureSessionPersistor, error) {
	mp := &MeasureSessionPersistor{
		session:           session,
		persist_frequency: 1 * time.Second,
	}
	if session.ParentBenchmark.CsvOuptut {
		DEBUG.Println("Initializing CSV")
		return mp, mp.init_csv()
	}
	return mp, nil
}
func (s *MeasureSessionPersistor) init_csv() error {
	name := s.session.Name
	err := os.Mkdir(name, os.ModeDir)
	if err != nil {
		if os.IsExist(err) {
			INFO.Printf("directory %s exists,", name)
			s.session.Name = name + time.Now().Format("_15:04:05")
			INFO.Printf("storing csv measures in directory %s \n", name)
			err = os.Mkdir(name, os.ModeDir)
			if err != nil {
				return fmt.Errorf("persistMeasures: %w", err)
			}
		}
	}
	s.csv = &csvPersistorData{}
	s.csv.PacketFile, err = os.Create(filepath.Join(name, filepath.Base("measure_packet.csv")))
	if err != nil {
		return fmt.Errorf("persistMeasures: %w", err)
	}

	s.csv.PacketWriter = csv.NewWriter(s.csv.PacketFile)
	heading := assets.CONST_HEADING
	if err := s.csv.PacketWriter.Write(heading); err != nil {
		return fmt.Errorf("persistMeasures: %w", err)
	}

	s.csv.QueueFile, err = os.Create(filepath.Join(name, filepath.Base("measure_queue.csv")))
	if err != nil {
		return fmt.Errorf("persistMeasures: %w", err)
	}

	s.csv.QueueWriter = csv.NewWriter(s.csv.QueueFile)
	heading = []string{"timestampMs", "memUsageBytes", "packetsinqueue"}
	if err := s.csv.QueueWriter.Write(heading); err != nil {
		return fmt.Errorf("persistMeasures: %w", err)
	}
	return nil
}
func (s *MeasureSessionPersistor) Close() {

	DEBUG.Println("Closing MeasureSessionPersistor")
	if s.csv != nil {
		s.csv.QueueWriter.Flush()
		s.csv.PacketWriter.Flush()
		s.csv.QueueFile.Close()
		s.csv.PacketFile.Close()
	}
}
func (s *MeasureSessionPersistor) persist(sample interface{}, r util.RoutineReport) {
	if err := (*s.db).Persist(sample); err != nil {
		r.Send_error_c <- fmt.Errorf("persisting sample (%+v) to DB: %w", sample, err)
	}
	if s.csv == nil {
		return
	}
	if csvsample, ok := sample.(datatypes.DB_measure_queue); ok {
		//DEBUG.Printf("Persisitng DB_measure_queue: %v", csvsample)
		if err := s.csv.QueueWriter.Write((csvsample).CsvRecord()); err != nil {
			r.Send_error_c <- fmt.Errorf("persisting DB_measure_queue (%+v) to csv file: %w", sample, err)
		}
		return
	}
	if csvsample, ok := sample.(datatypes.DB_measure_packet); ok {
		//DEBUG.Printf("Persisitng DB_measure_packet: %v", csvsample)
		if err := s.csv.PacketWriter.Write((csvsample).CsvRecord()); err != nil {
			r.Send_error_c <- fmt.Errorf("persisting DB_measure_packet (%+v) to csv file: %w", sample, err)
		}
	}
}

func (s *MeasureSessionPersistor) Run(r util.RoutineReport, samples chan interface{}) {
	//Setup
	var err error
	s.db, err = persistence.GetPersistence()
	if err != nil {
		r.Send_error_c <- fmt.Errorf("persistMeasures: %w", err)
	}
	tickerPersist := time.NewTicker(s.persist_frequency)
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
						s.persist(sampleInterface, r)
						avgLoadKbits += sample.LoadKbits
						samplePacketCount++
					case DB_measure_queue:
						s.persist(sampleInterface, r)
						sampleQueueCount++
					case exitStruct:
						DEBUG.Println("Exiting persistMeasures: due to struct")
						r.Wg.Done()
						return
					default:
						FATAL.Printf("Unexpected Input in persistMeasures %+v", sampleInterface)
					}
				default:
					readSamples = false
				}
			}
			(*s.db).Commit()
			//to csv
			if s.csv != nil {
				s.csv.PacketWriter.Flush()
				s.csv.QueueWriter.Flush()
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
