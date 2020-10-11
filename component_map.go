package akara

// ComponentMap is the interface for a component map; something that can look up
// components from an entity ID.
type ComponentMap interface {
	Component
	Init(*World)               // sets up
	Add(EID) Component         // returns the component that was added for the given entity
	Get(EID) (Component, bool) // returns the component and bool if component exists
	Remove(EID)                // removes the component for the given entity
}

// ComponentMapProvider is something that can be used to create a new component map instance
type ComponentMapProvider interface {
	NewMap() ComponentMap
}
