package functions

type function praedicates

func (f function) len() int                                  { return praedicates(f).Len() }
func (f function) accs() []Parametric                        { return praedicates(f).Accs() }
func (f function) pairs() []Paired                           { return praedicates(f).Pairs() }
func (f function) apply(p ...Paired) ([]Paired, Praedicates) { return praedicates(f).Apply(p...) }
