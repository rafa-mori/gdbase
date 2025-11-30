// Package adaptive provides interfaces and types for adaptive services.
package adaptive

type Adaptive interface {
	IsMongo(v any) bool
}
