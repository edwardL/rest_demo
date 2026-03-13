package event

import "sync"

type EventBusChan struct {
	mux      sync.RWMutex
	channels map[string][]chan IEvent
}

func NewEventBusChan() *EventBusChan {
	return &EventBusChan{
		mux:      sync.RWMutex{},
		channels: make(map[string][]chan IEvent),
	}
}

func (eb *EventBusChan) Subscribe(tp string, maxLen ...int) <-chan IEvent {
	eb.mux.Lock()
	defer eb.mux.Unlock()
	l := 1
	if len(maxLen) > 0 && maxLen[0] > 0 {
		l = maxLen[0]
	}
	ch := make(chan IEvent, l)
	if v, ok := eb.channels[tp]; ok {
		eb.channels[tp] = append(v, ch)
	} else {
		eb.channels[tp] = []chan IEvent{ch}
	}
	return ch
}

func (eb *EventBusChan) UnSubscribe(tp string, ch <-chan IEvent) {
	eb.mux.Lock()
	defer eb.mux.Unlock()
	if v, ok := eb.channels[tp]; ok {
		for i := range v {
			currCh := v[i]
			if currCh == ch {
				eb.channels[tp] = append(v[0:i], v[i+1:]...)
				close(currCh)
				return
			}
		}
	}
}

func (eb *EventBusChan) Publish(evt IEvent) {
	eb.mux.RLock()
	channels := make([]chan IEvent, 0)
	if v, ok := eb.channels[evt.Topic()]; ok {
		channels = append(channels, v...)
	}
	eb.mux.RUnlock()
	if len(channels) == 0 {
		return
	}
	for i := range channels {
		go func(ch chan IEvent) {
			ch <- evt
		}(channels[i])
	}
}
