package redis

import (
	"fmt"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	rdb, err := NewRedis(&Config{
		Host:     "192.168.50.60",
		Port:     6379,
		Password: "edward",
	})
	if err != nil {
		t.Error(err)
	}
	l := NewRateLimit(RateLimiterConfig{
		Limit:       5,
		LimitWindow: time.Second * 5,
	}, rdb)
	for i := 0; i < 200; i++ {
		fmt.Println(l.Allow("ttt"))
		time.Sleep(time.Millisecond * 1000)
	}
}
