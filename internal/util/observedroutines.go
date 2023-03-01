package util

import "sync"

type ErrorLevel uint8

const (
	ErrInfo ErrorLevel = iota
	ErrWarn
	ErrFatal
)

type RoutineReport struct {
	Wg               *sync.WaitGroup
	On_extern_exit_c chan uint8
	//This channel should be used, in the event of a fatal-ish error
	Send_error_c chan struct {
		Err   error
		Level ErrorLevel
	}
	//This channel should be used, if and only if some goroutine markes the application as finished
	Application_has_finished chan string
	exits                    []chan interface{}
}

func (r RoutineReport) Report(err error, level ErrorLevel) {
	r.Send_error_c <- struct {
		Err   error
		Level ErrorLevel
	}{
		Err:   err,
		Level: level,
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
