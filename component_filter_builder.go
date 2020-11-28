package akara

// ComponentFilterBuilder creates a component filter config
type ComponentFilterBuilder struct {
	world      *World
	require    []Component
	requireOne []Component
	forbid     []Component
}

func (cfb *ComponentFilterBuilder) Build() *ComponentFilter {
	bs := NewBitSet()
	f := &ComponentFilter{}

	for idx := range cfb.require {
		componentID := cfb.world.RegisterComponent(cfb.require[idx])

		bs.Set(int(componentID), true)
	}

	f.Required = bs.Clone()
	bs.Clear()

	for idx := range cfb.requireOne {
		componentID := cfb.world.RegisterComponent(cfb.requireOne[idx])

		bs.Set(int(componentID), true)
	}

	f.OneRequired = bs.Clone()
	bs.Clear()

	for idx := range cfb.forbid {
		componentID := cfb.world.RegisterComponent(cfb.forbid[idx])

		bs.Set(int(componentID), true)
	}

	f.Forbidden = bs

	return f
}

// Require makes all of the given components required by the filter
func (cfb *ComponentFilterBuilder) Require(components ...Component) *ComponentFilterBuilder {
	for idx := range components {
		cfb.require = append(cfb.require, components[idx])
	}

	return cfb
}

// RequireOne makes at least one of the given components required
func (cfb *ComponentFilterBuilder) RequireOne(components ...Component) *ComponentFilterBuilder {
	for idx := range components {
		cfb.requireOne = append(cfb.requireOne, components[idx])
	}

	return cfb
}

// Forbid makes all of the given components forbidden by the filter
func (cfb *ComponentFilterBuilder) Forbid(components ...Component) *ComponentFilterBuilder {
	for idx := range components {
		cfb.forbid = append(cfb.forbid, components[idx])
	}

	return cfb
}
