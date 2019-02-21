package run

import (
	p "github.com/JoergReinhardt/gatwd/parse"
)

type State interface {
	Run()
}
type StateFnc func() StateFnc

func (s StateFnc) Run() {
	for state := s(); state != nil; {
		state = state()
	}
}

type Queue interface {
	HasToken() bool
	Put(p.Token)
	Pull() p.Token
	Peek() p.Token
	PeekN(n int) p.Token
}
