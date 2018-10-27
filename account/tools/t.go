package main

import (
	"fmt"
	"os"
	"os/signal"
)

func Wait() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}

func main() {
	defer func() {
		fmt.Println("called")
	}()

	Wait()
}
