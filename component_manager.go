package akara

import (
	"fmt"
)

// NewComponentManager creates a new component manager
func NewComponentManager(w *World) *ComponentManager {
	cm := &ComponentManager{
		world: w,
		maps:  make(map[ComponentID]ComponentMap),
	}

	return cm
}

// ComponentManager maintains all of the component maps for all component types
type ComponentManager struct {
	world *World
	maps  map[ComponentID]ComponentMap
}

// InjectMap attempts to add the component map and returns whatever the component manager
// has associated with that component id
func (cm *ComponentManager) InjectMap(c Component) ComponentMap {
	cm.AddMap(c)
	return cm.maps[c.ID()]
}

// AddMap adds a component map to the component manager maps
func (cm *ComponentManager) AddMap(c Component) {
	if !cm.ContainsMap(c) {
		cm.maps[c.ID()] = c.NewMap()
		cm.maps[c.ID()].Init(cm.world)
	}
}

// GetMap returns the map for the given component ID
func (cm *ComponentManager) GetMap(c Component) (ComponentMap, error) {
	if cm.ContainsMap(c) {
		return cm.maps[c.ID()], nil
	}

	return nil, fmt.Errorf("no map found for component ID %d", c.ID())
}

// ContainsMap checks if the component manager has the given component map
// nolint:interfacer // components and component maps are the same for identification purposes
func (cm *ComponentManager) ContainsMap(c Component) bool {
	_, found := cm.maps[c.ID()]

	return found
}

// RemoveEntity removes the entity from all components maps
func (cm *ComponentManager) RemoveEntity(id EID) {
	for idx := range cm.maps {
		cm.maps[idx].Remove(id)
	}
}
