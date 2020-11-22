package akara

type comProvider = func() Component
type mapProvider = func() ComponentMap

func NewBaseComponent(id ComponentID, comFn comProvider, mapFn mapProvider) *BaseComponent {
	return &BaseComponent{
		id:          id,
		comProvider: comFn,
		mapProvider: mapFn,
	}
}

type BaseComponent struct {
	id ComponentID
	comProvider
	mapProvider
}

func (b *BaseComponent) ID() ComponentID {
	return b.id
}

func (b *BaseComponent) NewComponent() Component {
	return b.comProvider()
}

func (b *BaseComponent) NewMap() ComponentMap {
	return b.mapProvider()
}
