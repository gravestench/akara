package akara

func newComponentFactory(id ComponentID) *ComponentFactory {
	return &ComponentFactory{
		id: id,
	}
}

type ComponentFactory struct {
	*ComponentMap
	world    *World
	id       ComponentID
	provider func() Component
}

func (b *ComponentFactory) ID() ComponentID {
	return b.id
}

func (b *ComponentFactory) factoryNew(id EID) Component {
	instance := b.provider()

	b.instances[id] = instance

	return instance
}
