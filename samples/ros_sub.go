package main

import (
	"fmt"
	"rosgo/comm"
	"time"
)

func main() {
	chatter_chan := make(chan string)
	comm.Subscribe("/chatter", chatter_chan)
	fmt.Printf("subscribed. waiting 10 seconds\n")
	go func() {
		for {
			msg := <-chatter_chan
			fmt.Printf("Received a string message: '%s'\n", msg)
		}
	}()
	time.Sleep(10*time.Second)
	fmt.Printf("done waiting\n")
}
