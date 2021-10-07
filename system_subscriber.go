package akara

type subscriptions = map[*ComponentFilter]*Subscription

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
	BaseSystem
	Subscriptions subscriptions
}

func (b *BaseSubscriberSystem) Base() System {
	return &b.BaseSystem
}

func (b *BaseSubscriberSystem) AddSubscription(s interface{}) *Subscription {
	var filter *ComponentFilter

	switch instance := s.(type) {
	case *Subscription:
		filter = instance.Filter
	case *ComponentFilter:
		filter = instance
	case *ComponentFilterBuilder:
		filter = instance.Build()
	default:
		return nil
	}

	if b.Subscriptions == nil {
		b.Subscriptions = make(subscriptions)
	}

	sub := b.World.AddSubscription(filter)
	b.Subscriptions[sub.Filter] = sub

	return sub
}

func (b *BaseSubscriberSystem) RemoveSubscription(sub *Subscription) {
	delete(b.Subscriptions,  sub.Filter)
}

func (b *BaseSubscriberSystem) GetSubscriptions() []*Subscription {
	result := make([]*Subscription, 0)

	for idx := range b.Subscriptions {
		result = append(result, b.Subscriptions[idx])
	}

	return result
}

