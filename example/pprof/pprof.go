package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

func readMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	log.Printf("===> Alloc %d Bytes HeapIdle %d Bytes, HeapReleased %d Bytes", ms.Alloc, ms.HeapIdle, ms.HeapReleased)
}

func test() {
	container := make([]int, 8)
	log.Println("====> loop begin.")
	for i := 0; i < 32*1000*1000; i++ {
		container = append(container, i)
		if i == 16*1000*1000 {
			readMemStats()
		}
	}
	log.Println("====> loop end.")
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	}()
	readMemStats()
	test()
	readMemStats()
	log.Println(("[force gc]"))
	runtime.GC()

	readMemStats()

	go func() {
		for {
			readMemStats()
			time.Sleep(10 * time.Second)
		}
	}()

	time.Sleep(3600 * time.Second)
}
