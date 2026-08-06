package akara

import (
	"fmt"
	"reflect"
	"time"
)

// System advances one part of a simulation. World.Update invokes systems
// sequentially in registration order.
type System interface {
	Update(*World, time.Duration) error
}

// SystemFunc adapts a function to System.
type SystemFunc func(*World, time.Duration) error

func (function SystemFunc) Update(world *World, delta time.Duration) error {
	return function(world, delta)
}

// SystemID identifies one system registration in a world.
type SystemID uint64

type systemEntry struct {
	id     SystemID
	system System
}

// AddSystem appends system to the deterministic update order.
func (world *World) AddSystem(system System) (SystemID, error) {
	if system == nil || isNilSystem(system) {
		return 0, ErrNilSystem
	}
	world.systemsMu.Lock()
	defer world.systemsMu.Unlock()
	if world.closed.Load() {
		return 0, ErrWorldClosed
	}
	world.nextSystemID++
	if world.nextSystemID == 0 {
		return 0, ErrSystemIDExhausted
	}
	id := world.nextSystemID
	world.systems = append(world.systems, systemEntry{id: id, system: system})
	return id, nil
}

// RemoveSystem removes a registration and reports whether it existed. A system
// removed during Update will not run on the next update.
func (world *World) RemoveSystem(id SystemID) bool {
	world.systemsMu.Lock()
	defer world.systemsMu.Unlock()
	for index, entry := range world.systems {
		if entry.id == id {
			world.systems = append(world.systems[:index], world.systems[index+1:]...)
			return true
		}
	}
	return false
}

// Update runs a stable snapshot of systems sequentially. Concurrent Update calls
// are serialized.
func (world *World) Update(delta time.Duration) error {
	world.updateMu.Lock()
	defer world.updateMu.Unlock()
	if world.closed.Load() {
		return ErrWorldClosed
	}
	world.systemsMu.RLock()
	systems := append([]systemEntry(nil), world.systems...)
	world.systemsMu.RUnlock()
	for _, entry := range systems {
		if world.closed.Load() {
			return ErrWorldClosed
		}
		if err := entry.system.Update(world, delta); err != nil {
			return fmt.Errorf("akara: system %d update: %w", entry.id, err)
		}
	}
	return nil
}

func isNilSystem(system System) bool {
	value := reflect.ValueOf(system)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// Systems returns the registered IDs in update order.
func (world *World) Systems() []SystemID {
	world.systemsMu.RLock()
	ids := make([]SystemID, len(world.systems))
	for index, entry := range world.systems {
		ids[index] = entry.id
	}
	world.systemsMu.RUnlock()
	return ids
}
