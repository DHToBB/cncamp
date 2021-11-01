package main

import (
	"fmt"
	"runtime"
	"time"
)

func producer(ch chan<- int) {
	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		ch <- i
		fmt.Println("producer: ", i)
	}
	fmt.Println("stop to send data")
}

func consumer(ch <-chan int) {
	time.Sleep(20 * time.Second)
	fmt.Println("consumer begin to receive data")

	for {
		select {
		case data := <-ch:
			fmt.Println("consumer: ", data)
		}
	}
}

func main() {
	ch := make(chan int, 10)

	go producer(ch)
	go consumer(ch)

	//阻止进程退出
	for {
		runtime.Gosched()
	}

	//这种方式需要保证go协程不全部阻塞睡眠的时候有效
	//<-make(chan int)
}
