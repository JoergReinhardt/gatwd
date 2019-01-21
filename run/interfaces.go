package run

type StateFn func(State) StateFn
