package gatw

// Cons can either be flag, rank, or name
//func Cons(args ...Item) Item {
//	if len(args) > 0 { // return classes, or functions
//		// return constructor monad for chosen kind of category
//		if a, ok := args[0].(Acc); ok {
//			switch a {
//			case Ranked: // CATEGORY OF CONSTRUCTORS RANKED BY POSITION {{{
//
//				var ( // enclosed values
//					cons = make([]Item, 0, len(args))
//				)
//
//				// RETURN RANKED CONSTRUCTOR
//				return NarFnc(func(args ...Item) Item {
//					if len(args) > 0 {
//						// lookup by rank
//						if i, ok := args[0].(Id); ok {
//							return cons[i]
//						}
//						// define (every item instance but an id)
//						cons = append(cons, args[0])
//						// define multiple types
//						if len(args) > 1 {
//							return e.Cons(args[1:]...)
//						}
//						// return unit
//						return Id(0)
//					}
//					return Expr(cons)
//				})
//				//}}}
//			case Named: // CONSTRUCT CATEGORY OF CONSTRUCTORS IDENTIFIED BY NAME {{{
//
//				var ( // enclosed values
//					cons   = make([]Item, 0, len(args)-1)
//					names  = map[string]int{}
//					update = func(
//						sym Symbol,
//						fnc NarFnc,
//					) (
//						map[string]int,
//						[]Item,
//					) {
//						var id = len(cons)
//						names[string(sym)] = id
//						return names, append(cons, fnc)
//					}
//				)
//
//				// RETURN NAMED CONSTRUCTOR
//				return NarFnc(func(args ...Item) Item {
//					if len(args) > 0 {
//						// define from arg pair
//						if pair, ok := args[0].(PairFnc); ok {
//							//if  args[0].Type() {
//							if sym, ok := pair.Left().(Symbol); ok {
//								if fnc, ok := pair.Right().(NarFnc); ok {
//									names, cons = update(sym, fnc)
//									if len(args) > 2 {
//										return e.Cons(args[2:]...)
//									}
//									return e
//								}
//							}
//						}
//						// define from independent args
//						if len(args) > 1 {
//							if sym, ok := args[0].(Symbol); ok {
//								if fnc, ok := args[1].(NarFnc); ok {
//									names, cons = update(sym, fnc)
//									if len(args) > 2 {
//										return e.Cons(args[2:]...)
//									}
//									return e
//								}
//							}
//						}
//						// lookup by name
//						if n, ok := args[0].(Symbol); ok {
//							return cons[names[string(n)]]
//						}
//						// lookup by rank
//						if id, ok := args[0].(Id); ok {
//							if len(cons) > int(id) {
//								return cons[id]
//							}
//						}
//						// return unit
//						return Symbol("")
//					}
//					// return all constructors
//					return Expr(cons)
//				})
//				//}}}
//			case Flagged: // CATEGORY OF CONSTRUCTORS RANKED BY BIT FLAG VALUE {{{
//				var (
//					cons   = make([]Item, 0, len(args)-1)
//					names  = map[string]int{}
//					update = func(
//						flag Flag, fnc NarFnc,
//					) (
//						map[string]int,
//						[]Item,
//					) {
//						// constructor slice length matches flag rank → append as next flag
//						if bits.Len(uint(flag)) == len(cons) {
//							names[flag.Name()] = len(cons)
//							cons = append(cons, fnc)
//						}
//						// flag rank is lesser length of constructor slice → assign to rank matching flag
//						if bits.Len(uint(flag)) < len(cons) {
//							names[flag.Name()] = len(cons)
//							cons[bits.Len(uint(flag))] = fnc
//						}
//						// flag rank is greater constructor slice length → expand slice and append
//						if bits.Len(uint(flag)) > len(cons) {
//							// extend cons slice to match flag rank by length
//							cons = append(cons, make(
//								[]Item, 0, bits.Len(uint(flag))-1-len(cons),
//							)...)
//							names[flag.Name()] = len(cons)
//							cons = append(cons, fnc)
//						}
//						return names, cons
//					}
//				)
//				// RETURN CONSTRUCTOR
//				return NarFnc(func(args ...Item) Item {
//					if len(args) > 0 {
//						// define from pair of items
//						if pair, ok := args[0].(PairFnc); ok {
//							if flag, ok := pair.Left().(Flag); ok {
//								if fnc, ok := pair.Right().(NarFnc); ok {
//									names, cons = update(flag, fnc)
//									if len(args) > 2 {
//										return e.Cons(args[2:]...)
//									}
//									return e
//								}
//							}
//						}
//						// define from independent items
//						if len(args) > 1 {
//							// define from two flags
//							if flag, ok := args[0].(Flag); ok {
//								if fnc, ok := args[1].(NarFnc); ok {
//									names, cons = update(flag, fnc)
//									if len(args) > 2 {
//										return e.Cons(args[2:]...)
//									}
//									return e
//								}
//							}
//						}
//
//						// first argument is the accessor to lookup by
//						var arg = args[0]
//						// lookup by flag
//						if f, ok := arg.(Flag); ok {
//							if len(cons) > bits.Len(uint(f)) {
//								return cons[bits.Len(uint(f))]
//							}
//						}
//						// lookup by name
//						if n, ok := arg.(Symbol); ok {
//							if id, ok := names[string(n)]; ok {
//								if len(cons) > id {
//									return cons[id]
//								}
//							}
//						}
//						// lookup by rank
//						if id, ok := arg.(Id); ok {
//							if len(cons) > int(id) {
//								if len(cons) > int(id) {
//									return cons[int(id)]
//								}
//							}
//						}
//						// return unit
//						return Types
//					}
//					// return constructors
//					return Expr(cons)
//				}) //}}}
//			}
//		}
//	}
//	return e
//}

//func (e cat) String() string { return "" } }}}
//}}}
