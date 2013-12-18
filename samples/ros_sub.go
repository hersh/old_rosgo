package main

import (
	"fmt"
	"rosgo/comm"
	"time"
)

func main() {
	chatter_chan := make(chan string)
	_, err := comm.Subscribe("/chubbles", "/chatter", "std_msgs/String", chatter_chan)
	if err != nil {
		fmt.Printf("Error subscribing: %v", err)
		return
	}
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
