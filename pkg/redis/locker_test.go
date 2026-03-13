package redis_test

import (
	"fmt"
	"rest_demo/pkg/redis"
	"testing"
	"time"
)

func TestNewLocker_Lock(t *testing.T) {
	rdb, err := redis.NewRedis(&redis.Config{
		Host:     "192.168.50.60",
		Port:     6379,
		Password: "edward",
	})
	if err != nil {
		t.Error(err)
	}
	l := redis.NewLocker("test", rdb)
	err = l.Lock(time.Second)
	fmt.Println("Lock:", err)
	go func() {
		l1 := redis.NewLocker("test", rdb)
		fmt.Println("应该获取锁失败 ", l1.Lock(time.Second))
		time.Sleep(time.Second * 2)
		l2 := redis.NewLocker("test", rdb)
		fmt.Println("应该获取锁成功 ", l2.Lock(time.Second))
	}()
	time.Sleep(time.Millisecond * 800)
	l.UnLock()
	time.Sleep(time.Second * 5)
}

func TestNewLocker_TryLock(t *testing.T) {
	rdb, err := redis.NewRedis(&redis.Config{
		Host:     "192.168.50.60",
		Port:     6379,
		Password: "edward",
	})
	if err != nil {
		t.Error(err)
	}
	l := redis.NewLocker("test", rdb)
	err = l.TryLock(time.Second)
	fmt.Println("TryLock:", err)
	go func() {
		l1 := redis.NewLocker("test", rdb)
		fmt.Println("应该获取锁失败 ", l1.TryLock(time.Second))
		time.Sleep(time.Second)
		fmt.Println("应该获取锁成功 ", l1.TryLock(time.Second))
	}()
	time.Sleep(time.Second * 2)
	l.UnLock()
	time.Sleep(time.Second * 5)
}
