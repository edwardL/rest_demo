package lock

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	var sem = NewSemaphore(3)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem.Acquire()
			fmt.Printf("Goroutine %d is running\n", id)
			time.Sleep(5 * time.Second)
			fmt.Printf("Goroutine %d is done\n", id)
			sem.Release()
		}(i)
	}
	wg.Wait()
	fmt.Println("All goroutines are done.")
}
