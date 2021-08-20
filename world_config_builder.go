package akara

// NewWorldConfig creates a world config builder instance
func NewWorldConfig() *WorldConfig {
	return &WorldConfig{
		systems:    make([]System, 0),
		components: make([]Component, 0),
	}
}

// WorldConfig is used to declare Systems and component mappers.
// This is to be passed to a World factory function.
type WorldConfig struct {
	systems    []System
	components []Component
}

// With is used to add either Systems or component maps.
//
// Examples:
// 	builder.With(&system{})
// 	builder.With(&component{})
//
//	builder.
//		With(&movementSystem{}).
//		With(&velocity{}).
//		With(&position{})
func (b *WorldConfig) With(arg interface{}) *WorldConfig {
	switch entry := arg.(type) {
	case System:
		b.systems = append(b.systems, entry)
	case Component:
		b.components = append(b.components, entry)
	}

	return b
}
