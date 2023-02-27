package measuresession

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/persistence"
	"github.com/telekom/aml-jens/internal/persistence/datatypes"
	"github.com/telekom/aml-jens/internal/util"
)

type MeasureSessionPersistor struct {
	session           *datatypes.DB_session
	persist_frequency time.Duration
	db                *persistence.Persistence
	// internal, anonymous, representation of needed io objects
	csv *struct {
		PacketFile   *os.File
		PacketWriter *csv.Writer
		QueueFile    *os.File
		QueueWriter  *csv.Writer
	}
}

// Creates a new Persistor object, bound to the current session.
// Will also initialize csv writers if session.parentbenchmark.csvoutput is true
//
// Defaults to a persitence frequency of 1 Second
func NewMeasureSessionPersistor(session *datatypes.DB_session) (*MeasureSessionPersistor, error) {
	mp := &MeasureSessionPersistor{
		session:           session,
		persist_frequency: 1 * time.Second,
	}
	if session.ParentBenchmark.PrintToStdOut {
		fmt.Println(strings.Join(assets.CONST_HEADING, " "))
	}
	if session.ParentBenchmark.CsvOuptut {
		DEBUG.Println("Initializing CSV")
		return mp, mp.init_csv()
	}
	return mp, nil
}

// internal: creats needed io objects for csv writing
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
	s.csv = &struct {
		PacketFile   *os.File
		PacketWriter *csv.Writer
		QueueFile    *os.File
		QueueWriter  *csv.Writer
	}{}
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

// Only has an effect if csv writing is active
// Flushes all Wirters and Closes all opened files.
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
	if measure_packet, ok := sample.(DB_measure_packet); ok && s.session.ParentBenchmark.PrintToStdOut {
		if err := measure_packet.PrintLine(); err != nil {
			WARN.Println("Could not write Measurement")
			r.Send_error_c <- err
			//Todo check wg behaviour
			//r.Wg.Done()
			return
		}
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
		return
	}
	INFO.Printf("MeasureSessionPersistor.persist(%+v): unknown behavior", sample)
}

// Stars Loop, which persits any samples comin in throug the samples channel
//
// Blocking, releases Wg
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
			DEBUG.Println("Closing persistor outer")
			r.Wg.Done()
			return
		case <-tickerPersist.C: //start := time.Now()

			readSamples := true
			for readSamples {
				select {
				case sampleInterface := <-samples:
					switch sampleInterface.(type) {
					// write measure to db
					case DB_measure_packet, DB_measure_queue:
						s.persist(sampleInterface, r)
					default:
						WARN.Printf("Unexpected Input in persistMeasures %+v", sampleInterface)
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

		}
	}
}
