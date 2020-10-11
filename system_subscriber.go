package akara

// NewSubscriberSystem  creates a new subscriber system instance from the given component filters
func NewSubscriberSystem(filters ...*ComponentFilter) *SubscriberSystem {
	ss := &SubscriberSystem{
		BaseSystem:    &BaseSystem{},
		Subscriptions: make([]*Subscription, len(filters)),
	}

	for idx := range filters {
		ss.Subscriptions[idx] = NewSubscription(filters[idx])
	}

	return ss
}

// SubscriberSystem is a system which uses subscriptions to iterate over entities.
// Subscriptions are maintained and updated by the entity manager in conjunction with
// the component maps (entities are added to subscriptions when they meet the filter requirements)
type SubscriberSystem struct {
	*BaseSystem
	Subscriptions []*Subscription
}
