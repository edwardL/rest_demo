package event

type Event struct {
	topic   string
	payload any
}

func NewEvent(topic string, payload any) *Event {
	return &Event{topic: topic, payload: payload}
}

func (e *Event) Topic() string {
	return e.topic
}

func (e *Event) Payload() any {
	return e.payload
}
