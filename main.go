package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	stateFile            = "/tmp/pomogori.state"
	workDurationMinutes  = 25
	breakDurationMinutes = 5
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pomogori <command> [options]")
		fmt.Println("Commands: work, break, status, stop")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "work":
		startCmd := flag.NewFlagSet("work", flag.ExitOnError)
		duration := startCmd.Int("d", workDurationMinutes, "duration in minutes")
		startCmd.Parse(os.Args[2:])
		start("work", *duration*60)

	case "break":
		breakCmd := flag.NewFlagSet("break", flag.ExitOnError)
		duration := breakCmd.Int("d", breakDurationMinutes, "duration in minutes")
		breakCmd.Parse(os.Args[2:])
		start("break", *duration*60)

	case "status":
		status()

	case "stop":
		stop()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func start(sessionType string, duration int) {
	now := time.Now().Unix()
	// Store: startTime|duration|sessionType
	state := fmt.Sprintf("%d|%d|%s", now, duration, sessionType)
	os.WriteFile(stateFile, []byte(state), 0644)
	fmt.Printf("%s session started: %s\n", strings.Title(sessionType), formatDuration(duration))
}

func status() {
	startTime, duration, sessionType, err := readState()
	if err != nil {
		fmt.Println("No timer running")
		return
	}

	elapsed := int(time.Now().Unix() - startTime)
	remaining := duration - elapsed

	if remaining <= 0 {
		fmt.Printf("%s session complete!\n", strings.Title(sessionType))
		return
	}

	fmt.Printf("%s: %s remaining\n", strings.Title(sessionType), formatDuration(remaining))
}

func stop() {
	err := os.Remove(stateFile)
	if err != nil {
		fmt.Println("No timer running")
		return
	}
	fmt.Println("Timer stopped")
}

func readState() (startTime int64, duration int, sessionType string, err error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return 0, 0, "", err
	}

	parts := strings.Split(string(data), "|")
	if len(parts) != 3 {
		return 0, 0, "", fmt.Errorf("invalid state format")
	}

	startTime, _ = strconv.ParseInt(parts[0], 10, 64)
	duration, _ = strconv.Atoi(parts[1])
	sessionType = parts[2]
	return startTime, duration, sessionType, nil
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
