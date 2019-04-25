/*
 */
package functions

import ()

type (
	CaseExpr   func(...Callable) (CaseExpr, Consumeable)
	SwitchExpr func(...Callable) Callable
)
