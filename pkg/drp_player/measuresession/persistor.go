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
	exit_persistor    chan uint8
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
		exit_persistor:    make(chan uint8),
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
func (s *MeasureSessionPersistor) close() {
	DEBUG.Println("Closing MeasureSessionPersistor")
	if s.csv != nil {
		s.csv.QueueWriter.Flush()
		s.csv.PacketWriter.Flush()
		s.csv.QueueFile.Close()
		s.csv.PacketFile.Close()
	}
	(*s.db).Commit()
	close(s.exit_persistor)
}
func (s *MeasureSessionPersistor) persist(sample interface{}) error {
	if err := (*s.db).Persist(sample); err != nil {
		//FATAL!
		return fmt.Errorf("persisting sample (%+v) to DB: %w", sample, err)
	}
	if measure_packet, ok := sample.(DB_measure_packet); ok && s.session.ParentBenchmark.PrintToStdOut {
		if err := measure_packet.PrintLine(); err != nil {
			return fmt.Errorf("could not write Measurement: %w", err)
		}
	}
	if s.csv == nil {
		return nil
	}
	if csvsample, ok := sample.(*datatypes.DB_measure_queue); ok {
		//DEBUG.Printf("Persisitng DB_measure_queue: %v", csvsample)
		if err := s.csv.QueueWriter.Write((csvsample).CsvRecord()); err != nil {
			return fmt.Errorf("persisting DB_measure_queue (%+v) to csv file: %w", sample, err)
		}
		return nil
	}
	if csvsample, ok := sample.(datatypes.DB_measure_packet); ok {
		//DEBUG.Printf("Persisitng DB_measure_packet: %v", csvsample)
		if err := s.csv.PacketWriter.Write((csvsample).CsvRecord()); err != nil {
			return fmt.Errorf("persisting DB_measure_packet (%+v) to csv file: %w", sample, err)
		}
		return nil
	}
	INFO.Printf("MeasureSessionPersistor.persist(%+v): unknown behavior", sample)
	return nil
}

// Stars Loop, which persits any samples coming in through the samples channel
//
// Blocking, releases Wg
func (s *MeasureSessionPersistor) Run(samples chan interface{}, report_error func(err error, lvl util.ErrorLevel), done func()) {
	//Setup
	var err error
	s.db, err = persistence.GetPersistence()
	if err != nil {
		report_error(fmt.Errorf("persistMeasures: %w", err), util.ErrWarn)
	}
	tickerPersist := time.NewTicker(s.persist_frequency)
	for {
		select {
		case <-s.exit_persistor:
			s.close()
			return
		case <-tickerPersist.C:
			DEBUG.Println("Ticker")
			readSamples := true
			for readSamples {
				select {
				case sampleInterface, ok := <-samples:
					if !ok {
						DEBUG.Println("Closing persistor due to closed channel")
						done()
						return
					}
					switch sample := sampleInterface.(type) {
					// write measure to db
					case DB_measure_packet:
						err := s.persist(sample)
						if err != nil {
							report_error(err, util.ErrFatal)
						}
					case DB_measure_queue:
						err := s.persist(&sample)
						if err != nil {
							report_error(err, util.ErrFatal)
						}
					default:
						report_error(
							fmt.Errorf("unexpected Input in persistMeasures %v", sampleInterface),
							util.ErrWarn)
					}
				default:
					readSamples = false
				}
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
