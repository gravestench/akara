package akara

import "sync"

func newComponentFactory(id ComponentID) *ComponentFactory {
	cf := &ComponentFactory{
		id: id,
		mux:       &sync.Mutex{},
		instances: make(map[EID]Component),
	}

	return cf
}

type ComponentFactory struct {
	world    *World
	id       ComponentID
	instances map[EID]Component
	provider func() Component
	mux *sync.Mutex
}

func (b *ComponentFactory) ID() ComponentID {
	return b.id
}

func (b *ComponentFactory) factoryNew(id EID) Component {
	instance := b.provider()

	b.instances[id] = instance

	return instance
}

func (m *ComponentFactory) Add(id EID) Component {
	if c, found := m.Get(id); found {
		return c
	}

	m.mux.Lock()

	c := m.factoryNew(id)

	m.mux.Unlock()

	m.world.UpdateEntity(id)

	return c
}

func (m *ComponentFactory) Get(id EID) (Component, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	c, found := m.instances[id]

	return c, found
}

func (m *ComponentFactory) Remove(id EID) {
	m.mux.Lock()

	delete(m.instances, id)

	m.mux.Unlock()

	m.world.UpdateEntity(id)
}
