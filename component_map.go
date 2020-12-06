package akara

import "sync"

// ComponentMap is used to keep track of component instances,
type ComponentMap struct {
	mux *sync.Mutex
	base *ComponentFactory
	instances map[EID]Component
}

func (m *ComponentMap) Add(id EID) Component {
	if c, found := m.Get(id); found {
		return c
	}

	m.mux.Lock()

	c := m.base.factoryNew(id)

	m.mux.Unlock()

	m.base.world.UpdateEntity(id)

	return c
}

func (m *ComponentMap) Get(id EID) (Component, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	c, found := m.instances[id]

	return c, found
}

func (m *ComponentMap) Remove(id EID) {
	m.mux.Lock()

	delete(m.instances, id)

	m.mux.Unlock()

	m.base.world.UpdateEntity(id)
}

