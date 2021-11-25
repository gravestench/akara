package akara

import "sync"

func newComponentFactory(id ComponentID) *ComponentFactory {
	cf := &ComponentFactory{
		id:        id,
		mux:       &sync.Mutex{},
		instances: make(map[EID]Component),
	}

	return cf
}

// ComponentFactory is used to manage a one-to-one mapping between entity ID's and
// component instances. The factory is responsible for creating, retrieving, and deleting
// components using a given component ID.
//
// Attempting to create more than one component for a given component ID will result
// in nothing happening (the existing component instance will still exist).
type ComponentFactory struct {
	world     *World
	id        ComponentID
	instances map[EID]Component
	provider  func() Component
	mux       *sync.Mutex
}

// ID returns the registered component ID for this component type
func (cf *ComponentFactory) ID() ComponentID {
	return cf.id
}

func (cf *ComponentFactory) factoryNew(id EID) Component {
	cf.instances[id] = cf.provider()

	return cf.instances[id]
}

// Add a new component for the given entity ID and yield the component.
// If a component already exists, yield the existing component.
//
// This operation will update world subscriptions for the given entity.
func (cf *ComponentFactory) Add(id EID) Component {
	if c, found := cf.Get(id); found {
		return c
	}

	cf.mux.Lock()

	c := cf.factoryNew(id)

	cf.mux.Unlock()

	cf.world.UpdateEntity(id)

	return c
}

// Get will yield the component and a bool, much like map retrieval.
// The bool indicates whether a component was found for the given entity ID.
// The component can be nil.
func (cf *ComponentFactory) Get(id EID) (Component, bool) {
	cf.mux.Lock()
	defer cf.mux.Unlock()

	c, found := cf.instances[id]

	return c, found
}

// Remove will destroy the component instance for the given entity ID.
// This operation will update world subscriptions for the given entity ID.
func (cf *ComponentFactory) Remove(id EID) {
	cf.mux.Lock()

	delete(cf.instances, id)

	cf.mux.Unlock()

	cf.world.UpdateEntity(id)
}
