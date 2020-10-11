package akara

// Component is something that can be identified with a component ID, as well
// as something that can create a new component map for itself
type Component interface {
	ComponentIdentifier
	ComponentMapProvider
}
