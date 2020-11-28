package akara

// Component can be anything with a `New` method that creates a new component instance
type Component interface {
	New() Component
}
