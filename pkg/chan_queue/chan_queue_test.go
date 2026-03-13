package chan_queue

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestChan(t *testing.T) {
	fn := func(ts int) {
		fmt.Println("handle start", ts)
		time.Sleep(time.Second * 3)
		fmt.Println("handle end", ts)
	}
	r := NewChanQueue(fn, SetChanQueueMaxNum[int](3), SetChanQueueLimit[int](3))
	go r.Start()
	for i := 0; i < 20; i++ {
		go r.Push(i)
	}
	time.Sleep(time.Second * 15)
}

func TestChanClose(t *testing.T) {
	fn := func(ts int) {
		fmt.Println("handle start", ts)
		time.Sleep(time.Second * 5)
		fmt.Println("handle end", ts)
	}
	r := NewChanQueue(fn, SetChanQueueMaxNum[int](6), SetChanQueueLimit[int](3), SetChanQueueGracefulClose[int](false))
	go r.Start()
	go func() {
		for i := 0; i < 20; i++ {
			fmt.Println(r.Push(i))
			time.Sleep(time.Second)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("close ", time.Now())
			r.Close()
			fmt.Println("close ", time.Now())
		}
	}()
	time.Sleep(time.Second * 35)
}
