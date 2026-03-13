package event

import "context"

type IEvent interface {
	Topic() string
	Payload() any
}

type IEventBus interface {
	Subscribe(string, ISubscriber)
	UnSubscribe(string, ISubscriber)
	Publish(IEvent)
	PublishAsync(IEvent)
}

type ISubscriber interface {
	Handle(IEvent)
}

type ISubscribeFunc func(IEvent)

func (f ISubscribeFunc) Handle(evt IEvent) {
	f(evt)
}

type IEventBusWithContext interface {
	SubscribeWithContext(string, ISubscriberWithContext)
	UnSubscribeWithContext(string, ISubscriberWithContext)
	PublishWithContext(context.Context, IEvent)
	PublishAsyncWithContext(context.Context, IEvent)
}

type ISubscriberWithContext interface {
	Handle(context.Context, IEvent)
}

type ISubscribeFuncWithContext func(context.Context, IEvent)

func (f ISubscribeFuncWithContext) Handle(ctx context.Context, evt IEvent) {
	f(ctx, evt)
}
