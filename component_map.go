package akara

// ComponentMap is used to keep track of component instances,
type ComponentMap struct {
	base *ComponentFactory
	instances map[EID]Component
}

func (m *ComponentMap) Add(id EID) Component {
	if c, found := m.Get(id); found {
		return c
	}

	c := m.base.factoryNew(id)

	m.base.world.UpdateEntity(id)

	return c
}

func (m *ComponentMap) Get(id EID) (Component, bool) {
	c, found := m.instances[id]
	return c, found
}

func (m *ComponentMap) Remove(id EID) {
	delete(m.instances, id)
}

