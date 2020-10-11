package akara

// ComponentID is a unique identifier for a component type
type ComponentID uint

// ComponentIdentifier is something which can be used to identify a component.
// This should be implemented by components, as well as component maps
type ComponentIdentifier interface {
	ID() ComponentID
}
