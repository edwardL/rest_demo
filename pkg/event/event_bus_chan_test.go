package event

import (
	"fmt"
	"testing"
	"time"
)

func TestEventBusChan(t *testing.T) {
	tp1 := "aaa"
	//tp2 := "bbb"
	eb := NewEventBusChan()
	ch1 := eb.Subscribe(tp1, 10)
	//ch2 := eb.Subscribe(tp2)
	ch3 := eb.Subscribe(tp1)
	go func() {
		for i := 0; i <= 20; i++ {
			//time.Sleep(time.Millisecond * 10)
			//if i == 2 {
			//eb.UnSubscribe(tp1, ch1)
			//eb.UnSubscribe(tp1, ch3)
			//eb.UnSubscribe(tp2, ch2)
			//}
			eb.Publish(NewEvent(tp1, i))
			//eb.Publish(NewEvent(tp2, i))
		}
	}()
	time.Sleep(time.Second)
	go func() {
		for v := range ch1 {
			time.Sleep(time.Millisecond * 1000)
			fmt.Println("ch1", v.Topic(), v.Payload(), len(ch1))
		}
	}()
	//go func() {
	//	for v := range ch2 {
	//		time.Sleep(time.Millisecond * 250)
	//		fmt.Println("ch2", v.Topic(), v.Payload(), len(ch2))
	//	}
	//}()
	go func() {
		for v := range ch3 {
			time.Sleep(time.Millisecond * 1000)
			fmt.Println("ch3", v.Topic(), v.Payload(), len(ch3))
		}
	}()
	//fmt.Println(len(ch1), len(ch3))
	time.Sleep(time.Second * 20)
}
