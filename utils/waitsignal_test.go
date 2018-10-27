package utils

import (
	"fmt"
	"testing"
	"time"
)

func Test_WaitKill(t *testing.T) {
	defer func() {
		fmt.Println("waitkill and exit was ok")
	}()
	fmt.Println("press ^c or wait 2 seconds")
	go func() {
		time.Sleep(time.Second * 2)
		Exit()
	}()
	WaitKill()
}
