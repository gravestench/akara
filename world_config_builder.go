package akara

// NewWorldConfig creates a world config builder instance
func NewWorldConfig() *WorldConfig {
	return &WorldConfig{
		systems:       make([]System, 0),
		componentMaps: make([]ComponentMap, 0),
	}
}

// WorldConfig is used to declare systems and component mappers.
// This is to be passed to a World factory function.
type WorldConfig struct {
	systems       []System
	componentMaps []ComponentMap
}

// With is used to add either systems or component maps.
//
// Examples:
// 	builder.With(system)
// 	builder.With(componentMap)
//
//	builder.
//		With(movementSystem).
//		With(velocityMap).
//		With(positionMap)
func (b *WorldConfig) With(arg interface{}) *WorldConfig {
	switch entry := arg.(type) {
	case System:
		b.systems = append(b.systems, entry)
	case ComponentMap:
		b.componentMaps = append(b.componentMaps, entry)
	}

	return b
}
