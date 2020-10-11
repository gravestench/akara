package akara

// NewFilter is shorthand for the longer function name
func NewFilter() *ComponentFilterBuilder {
	return NewComponentFilterBuilder()
}

// NewComponentFilterBuilder creates a builder for creating
func NewComponentFilterBuilder() *ComponentFilterBuilder {
	return &ComponentFilterBuilder{
		require:    make([]ComponentIdentifier, 0),
		requireOne: make([]ComponentIdentifier, 0),
		forbid:     make([]ComponentIdentifier, 0),
	}
}

// ComponentFilterBuilder creates a component filter config
type ComponentFilterBuilder struct {
	require    []ComponentIdentifier
	requireOne []ComponentIdentifier
	forbid     []ComponentIdentifier
}

func (cfb *ComponentFilterBuilder) Build() *ComponentFilter {
	bs := NewBitSet()
	f := &ComponentFilter{}

	for idx := range cfb.require {
		bs.Set(int(cfb.require[idx].ID()), true)
	}

	f.Required = bs.Clone()
	bs.Clear()

	for idx := range cfb.requireOne {
		bs.Set(int(cfb.requireOne[idx].ID()), true)
	}

	f.OneRequired = bs.Clone()
	bs.Clear()

	for idx := range cfb.forbid {
		bs.Set(int(cfb.forbid[idx].ID()), true)
	}

	f.Forbidden = bs

	return f
}

// Require makes all of the given components required by the filter
func (cfb *ComponentFilterBuilder) Require(components ...ComponentIdentifier) *ComponentFilterBuilder {
	for idx := range components {
		cfb.require = append(cfb.require, components[idx])
	}

	return cfb
}

// RequireOne makes at least one of the given components required
func (cfb *ComponentFilterBuilder) RequireOne(components ...ComponentIdentifier) *ComponentFilterBuilder {
	for idx := range components {
		cfb.requireOne = append(cfb.requireOne, components[idx])
	}

	return cfb
}

// Forbid makes all of the given components forbidden by the filter
func (cfb *ComponentFilterBuilder) Forbid(components ...ComponentIdentifier) *ComponentFilterBuilder {
	for idx := range components {
		cfb.forbid = append(cfb.forbid, components[idx])
	}

	return cfb
}
