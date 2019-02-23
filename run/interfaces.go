package run

import ()

type State interface {
	Run()
}
type StateFnc func() StateFnc

func (s StateFnc) Run() {
	for state := s(); state != nil; {
		state = state()
	}
}
