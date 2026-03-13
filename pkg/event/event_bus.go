package event

import (
	"context"
	"log"
	"sync"
)

type EventBus struct {
	mux                    sync.RWMutex
	subscribers            map[string][]ISubscriber
	subscribersWithContext map[string][]ISubscriberWithContext
}

func NewEventBus() *EventBus {
	return &EventBus{
		mux:                    sync.RWMutex{},
		subscribers:            make(map[string][]ISubscriber),
		subscribersWithContext: make(map[string][]ISubscriberWithContext),
	}
}

func (eb *EventBus) Subscribe(tp string, sub ISubscriber) {
	eb.mux.Lock()
	defer eb.mux.Unlock()
	if v, ok := eb.subscribers[tp]; ok {
		eb.subscribers[tp] = append(v, sub)
	} else {
		eb.subscribers[tp] = []ISubscriber{sub}
	}
}

func (eb *EventBus) UnSubscribe(tp string, sub ISubscriber) {
	eb.mux.Lock()
	defer eb.mux.Unlock()
	if v, ok := eb.subscribers[tp]; ok {
		for i := range v {
			if v[i] == sub {
				eb.subscribers[tp] = append(v[0:i], v[i+1:]...)
				return
			}
		}
	}
}

func (eb *EventBus) Publish(evt IEvent) {
	eb.mux.RLock()
	subscribes := make([]ISubscriber, 0)
	if v, ok := eb.subscribers[evt.Topic()]; ok {
		subscribes = append(subscribes, v...)
	}
	eb.mux.RUnlock()
	if len(subscribes) == 0 {
		return
	}
	for _, sub := range subscribes {
		sub.Handle(evt)
	}
}

func (eb *EventBus) PublishAsync(evt IEvent) {
	eb.mux.RLock()
	subscribers := make([]ISubscriber, 0)
	if v, ok := eb.subscribers[evt.Topic()]; ok {
		subscribers = append(subscribers, v...)
	}
	eb.mux.RUnlock()
	if len(subscribers) == 0 {
		return
	}

	for i := range subscribers {
		go func(handle ISubscriber) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("event[%s] dispatchAsync panic:%s", evt.Topic(), err)
				}
			}()
		}(subscribers[i])
	}
}

func (eb *EventBus) SubscribeWithContext(tp string, sub ISubscriberWithContext) {
	eb.mux.Lock()
	defer eb.mux.Unlock()
	if v, ok := eb.subscribersWithContext[tp]; ok {
		eb.subscribersWithContext[tp] = append(v, sub)
	} else {
		eb.subscribersWithContext[tp] = []ISubscriberWithContext{sub}
	}
}

func (eb *EventBus) UnSubscribeWithContext(tp string, sub ISubscriberWithContext) {
	eb.mux.Lock()
	defer eb.mux.Unlock()
	if v, ok := eb.subscribersWithContext[tp]; ok {
		for i := range v {
			if v[i] == sub {
				eb.subscribersWithContext[tp] = append(v[0:i], v[i+1:]...)
				return
			}
		}
	}
}

func (eb *EventBus) PublishWithContext(ctx context.Context, evt IEvent) {
	eb.mux.RLock()
	subscribers := make([]ISubscriberWithContext, 0)
	if v, ok := eb.subscribersWithContext[evt.Topic()]; ok {
		subscribers = append(subscribers, v...)
	}
	eb.mux.RUnlock()
	if len(subscribers) == 0 {
		return
	}
	for _, sub := range subscribers {
		sub.Handle(ctx, evt)
	}
}

func (eb *EventBus) PublishAsyncWithContext(ctx context.Context, evt IEvent) {
	eb.mux.RLock()
	subscribers := make([]ISubscriberWithContext, 0)
	if v, ok := eb.subscribersWithContext[evt.Topic()]; ok {
		subscribers = append(subscribers, v...)
	}
	eb.mux.RUnlock()
	if len(subscribers) == 0 {
		return
	}
	for i := range subscribers {
		go func(handle ISubscriberWithContext) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("event[%s] dispatchAsync panic:%s", evt.Topic(), err)
				}
			}()
			handle.Handle(ctx, evt)
		}(subscribers[i])
	}
}
