package util

import (
	"sync"
	"time"
)

type ErrorLevel uint8

const (
	ErrInfo ErrorLevel = iota
	ErrWarn
	ErrFatal
)

type RoutineReport struct {
	Wg              *sync.WaitGroup
	Exit_now_signal chan uint8
	//This channel should be used, in the event of a fatal-ish error
	Send_error_c chan struct {
		Err   error
		Level ErrorLevel
	}
	//This channel should be used, if and only if some goroutine markes the application as finished
	Application_has_finished chan string
}

func (r RoutineReport) Report(err error, level ErrorLevel) {
	WARN.Printf("[%d]%+v\n", level, err)

	select {
	case r.Send_error_c <- struct {
		Err   error
		Level ErrorLevel
	}{
		Err:   err,
		Level: level,
	}:
	case <-time.After(100 * time.Millisecond):
		WARN.Printf("Could not report Error[%d]: %s", level, err)
	}
}
func (r RoutineReport) ReportFatal(err error) {
	r.Report(err, ErrFatal)
}
func (r RoutineReport) ReportWarn(err error) {
	r.Report(err, ErrWarn)
}
func (r RoutineReport) ReportInfo(err error) {
	r.Report(err, ErrInfo)
}
