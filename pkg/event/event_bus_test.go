package event

import (
	"fmt"
	"testing"
)

func TestEventBus(t *testing.T) {
	tp1 := "aaa"
	tp2 := "bbb"
	f1 := func(evt IEvent) {
		fmt.Println("f1:", evt.Topic(), evt.Payload())
	}
	f2 := func(evt IEvent) {
		fmt.Println("f2:", evt.Topic(), evt.Payload())
	}
	fn1 := ISubscribeFunc(f1)
	fn2 := ISubscribeFunc(f2)
	eb := NewEventBus()
	eb.Subscribe(tp1, &fn1)
	eb.Subscribe(tp2, &fn2)
	eb.Subscribe(tp1, &fn2)

	for i := 0; i <= 4; i++ {
		if i == 2 {
			eb.UnSubscribe(tp1, &fn2)
		}
		eb.Publish(NewEvent(tp1, i))
		eb.Publish(NewEvent(tp2, i))
		eb.PublishAsync(NewEvent(tp1, i))
		eb.PublishAsync(NewEvent(tp2, i))
	}
}
