package main

import (
	"fmt"
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
	arr := []string{"I", "am", "stupid", "and", "weak"}

	for k, _ := range arr {
		switch k {
		case 2:
			arr[k] = "smart"
		case 4:
			arr[k] = "strong"
		}
	}

	fmt.Printf("%+v\n", arr)
}
