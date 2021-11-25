package akara

import "github.com/gravestench/bitset"

// NewComponentFilterBuilder creates a new builder for a component filter, using
// the given world.
func NewComponentFilterBuilder(w *World) *ComponentFilterBuilder {
	return &ComponentFilterBuilder{
		world:      w,
		require:    make([]Component, 0),
		requireOne: make([]Component, 0),
		forbid:     make([]Component, 0),
	}
}

// ComponentFilterBuilder creates a component filter config
type ComponentFilterBuilder struct {
	world      *World
	require    []Component
	requireOne []Component
	forbid     []Component
}

// Build iterates through all components in the filter and registers them in the world,
// to ensure that all components have a unique Component ID. Then, the bitsets for
// Required, OneRequired, and Forbidden bits are set in the corresponding bitsets of the filter.
func (cfb *ComponentFilterBuilder) Build() *ComponentFilter {
	f := &ComponentFilter{
		Required:    bitset.NewBitSet(),
		OneRequired: bitset.NewBitSet(),
		Forbidden:   bitset.NewBitSet(),
	}

	for idx := range cfb.require {
		componentID := cfb.world.RegisterComponent(cfb.require[idx])

		f.Required.Set(int(componentID), true)
	}

	for idx := range cfb.requireOne {
		componentID := cfb.world.RegisterComponent(cfb.requireOne[idx])

		f.OneRequired.Set(int(componentID), true)
	}

	for idx := range cfb.forbid {
		componentID := cfb.world.RegisterComponent(cfb.forbid[idx])

		f.Forbidden.Set(int(componentID), true)
	}

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
