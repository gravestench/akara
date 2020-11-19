package akara

// NewBaseSubscriberSystem creates a new subscriber system instance from the given component filters
func NewBaseSubscriberSystem(filters ...*ComponentFilter) *BaseSubscriberSystem {
	ss := &BaseSubscriberSystem{
		BaseSystem:    &BaseSystem{},
	}

	for idx := range filters {
		ss.AddSubscription(NewSubscription(filters[idx]))
	}

	return ss
}

type SubscriberSystem interface {
	System
	hasBaseSystem
	AddSubscription(s *Subscription)
	RemoveSubscription(s *Subscription)
	GetSubscriptions() []*Subscription
}

// BaseSubscriberSystem is a system which uses subscriptions to iterate over entities.
// Subscriptions are maintained and updated by the entity manager in conjunction with
// the component maps (entities are added to subscriptions when they meet the filter requirements)
type BaseSubscriberSystem struct {
	*BaseSystem
	Subscriptions []*Subscription
}

func (b *BaseSubscriberSystem) Base() *BaseSystem {
	return b.BaseSystem
}

func (b *BaseSubscriberSystem) AddSubscription(sub *Subscription) {
	if b.Subscriptions == nil {
		b.Subscriptions = make([]*Subscription, 0)
	}

	b.Subscriptions = append(b.Subscriptions, sub)
}

func (b *BaseSubscriberSystem) RemoveSubscription(sub *Subscription) {
	for idx := range b.Subscriptions {
		if b.Subscriptions[idx] != sub {
			continue
		}

		b.Subscriptions = append(b.Subscriptions[:idx], b.Subscriptions[idx+1:]...)
	}
}

func (b *BaseSubscriberSystem) GetSubscriptions() []*Subscription {
	return b.Subscriptions
}

