package main

import (
	"fmt"
	"time"
)

type RepeatedTimer struct {
	interval  time.Duration
	function  func()
	stopChan  chan bool
	isRunning bool
}

func NewRepeatedTimer(interval time.Duration, function func()) *RepeatedTimer {
	rt := &RepeatedTimer{
		interval:  interval,
		function:  function,
		stopChan:  make(chan bool),
		isRunning: false,
	}
	rt.Start()
	return rt
}

func (rt *RepeatedTimer) Start() {
	if rt.isRunning {
		return
	}

	rt.isRunning = true
	go func() {
		ticker := time.NewTicker(rt.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rt.function()
			case <-rt.stopChan:
				return
			}
		}
	}()
}

func (rt *RepeatedTimer) Stop() {
	if !rt.isRunning {
		return
	}

	rt.isRunning = false
	rt.stopChan <- true
	close(rt.stopChan)
	rt.stopChan = make(chan bool) // Reset the channel for future use
}

func main() {
	// Example usage
	timer := NewRepeatedTimer(2*time.Second, func() {
		fmt.Println("Hello, every 2 seconds!")
	})

	// Let the timer run for 10 seconds
	time.Sleep(10 * time.Second)
	timer.Stop()
	fmt.Println("Timer stopped!")
}
