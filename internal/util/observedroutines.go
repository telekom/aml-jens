package util

import "sync"

type RoutineReport struct {
	Wg               *sync.WaitGroup
	On_extern_exit_c chan uint8
	//This channel should be used, in the event of a fatal-ish error
	Send_error_c chan error
	//This channel should be used, if and only if some goroutine markes the application as finished
	Application_has_finished chan string
}
