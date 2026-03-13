package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"time"
)

func main() {
	// 创建trace 文件
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 启动trace gorouting
	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()

	// main
	//fmt.Println("Hello World")
	fmt.Printf("%d", time.Now().UnixMilli())
}
