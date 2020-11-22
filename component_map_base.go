package akara

func NewBaseComponentMap(base *BaseComponent) *BaseComponentMap {
	return &BaseComponentMap{
		BaseComponent: base,
		instances: make(map[EID]Component),
	}
}

type BaseComponentMap struct {
	world *World
	*BaseComponent
	instances map[EID]Component
}

func (b *BaseComponentMap) Init(w *World) {
	b.world = w
}

func (b *BaseComponentMap) Add(id EID) Component {
	if com, has := b.instances[id]; has {
		return com
	}

	b.instances[id] = b.NewComponent()
	b.world.UpdateEntity(id)

	return b.instances[id]
}

func (b *BaseComponentMap) Get(id EID) (c Component, found bool) {
	entry, found := b.instances[id]
	return entry, found
}

func (b *BaseComponentMap) Remove(id EID) {
	delete(b.instances, id)
	b.world.UpdateEntity(id)
}
