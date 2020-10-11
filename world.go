package akara

import (
	"time"
)

// NewWorld creates a new world instance
func NewWorld(cfg *WorldConfig) *World {
	world := &World{
		Systems: make([]System, 0),
	}

	world.EntityManager = NewEntityManager(world)
	world.ComponentManager = NewComponentManager(world)

	for _, m := range cfg.componentMaps {
		world.ComponentManager.InjectMap(m)
	}

	for _, system := range cfg.systems {
		world.AddSystem(system)
	}

	return world
}

// World contains all of the Systems, as well as an EntityManager and ComponentManager
type World struct {
	TimeDelta time.Duration
	*ComponentManager
	*EntityManager
	Systems []System
}

// RemoveEntity removes an entity, and its components, from the world
func (w *World) RemoveEntity(id uint64) {
	w.EntityManager.RemoveEntity(id)
	w.ComponentManager.RemoveEntity(id)
}

// AddSystem adds a system to the world
func (w *World) AddSystem(s System) *World {
	w.Systems = append(w.Systems, s)

	s.SetActive(true)

	if initializer, ok := s.(SystemInitializer); ok {
		initializer.Init(w)
	}

	return w
}

// UpdateEntity updates the entity in the world. This causes the entity manager to
// update all subscriptions for this entity ID.
func (w *World) UpdateEntity(id EID) {
	w.UpdateSubscriptions(id)
}

// Update iterates through all Systems and calls the update method if the system is active
func (w *World) Update(dt time.Duration) error {
	w.TimeDelta = dt

	for sysIdx := range w.Systems {
		if w.Systems[sysIdx].Active() {
			w.Systems[sysIdx].Process()
		}
	}

	return nil
}
