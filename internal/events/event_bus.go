package events

import "sync"

type EventBus struct {
	events map[string]map[string][]func(...any)
	mutex  sync.Mutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		events: make(map[string]map[string][]func(...any)),
	}
}

// Adiciona eventos
func (e *EventBus) On(name, event string, callback func(...any)) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.events[name] == nil {
		e.events[name] = make(map[string][]func(...any))
	}

	e.events[name][event] = append(e.events[name][event], callback)
}

// Dispara eventos
func (e *EventBus) Emit(name, event string, args ...any) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if callbacks, ok := e.events[name][event]; ok {
		for _, cb := range callbacks {
			go cb(args...)
		}
	}
}
