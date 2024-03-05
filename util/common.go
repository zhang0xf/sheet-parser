package util

import (
	"math"
	"os"
	"os/signal"
	"syscall"
)

// f:需要处理的浮点数，n：要保留小数的位数
func RoundFloat(f float64, n int) float64 {
	n10 := math.Pow10(n)
	if f < 0 {
		return math.Trunc((math.Abs(f)+0.5/n10)*n10*-1) / n10
	} else {
		return math.Trunc((math.Abs(f)+0.5/n10)*n10) / n10
	}
}

func WaitTerminate() {
	exitChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		close(exitChan)
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-exitChan
}
