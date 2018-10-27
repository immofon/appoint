package utils

import (
	"os"
	"os/signal"
)

var signalChan chan os.Signal = make(chan os.Signal, 10)

func WaitKill() {
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	<-signalChan
}

func Exit() {
	go func() { signalChan <- os.Kill }()
}
