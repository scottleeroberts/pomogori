package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	stateFile       = "/tmp/pomogori.state"
	defaultDuration = 25 * 60 // 25 minutes in seconds
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pomogori [start|status|stop]")
		return
	}

	switch os.Args[1] {
	case "start":
		start()
	case "status":
		status()
	case "stop":
		stop()
	}
}

func start() {
	now := time.Now().Unix()
	// Store: startTime|duration
	state := fmt.Sprintf("%d|%d", now, defaultDuration)
	os.WriteFile(stateFile, []byte(state), 0644)
	fmt.Printf("Timer started: %s\n", formatDuration(defaultDuration))
}

func status() {
	startTime, duration, err := readState()
	if err != nil {
		fmt.Println("No timer running")
		return
	}

	elapsed := int(time.Now().Unix() - startTime)
	remaining := duration - elapsed

	if remaining <= 0 {
		fmt.Println("Timer complete!")
		return
	}

	fmt.Printf("Remaining: %s\n", formatDuration(remaining))
}

func stop() {
	err := os.Remove(stateFile)
	if err != nil {
		fmt.Println("No timer running")
		return
	}
	fmt.Println("Timer stopped")
}

func readState() (startTime int64, duration int, err error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(string(data), "|")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid state format")
	}

	startTime, _ = strconv.ParseInt(parts[0], 10, 64)
	duration, _ = strconv.Atoi(parts[1])
	return startTime, duration, nil
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
