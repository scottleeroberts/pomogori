package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	stateFile            = "/tmp/pomogori.state"
	workDurationMinutes  = 25
	breakDurationMinutes = 5
)

type state struct {
	startTime   int64
	duration    int
	sessionType string
	paused      bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pomogori <command> [options]")
		fmt.Println("Commands: work, break, pause, resume, status, watch, stop")
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

	case "pause":
		pause()

	case "resume":
		resume()

	case "stop":
		stop()

	case "watch":
		watch()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func start(sessionType string, duration int) {
	now := time.Now().Unix()
	// Store: startTime|duration|sessionType|status
	stateStr := fmt.Sprintf("%d|%d|%s|running", now, duration, sessionType)
	os.WriteFile(stateFile, []byte(stateStr), 0644)
	fmt.Printf("%s session started: %s\n", strings.Title(sessionType), formatDuration(duration))
}

func status() {
	s, err := readState()
	if err != nil {
		fmt.Println("No timer running")
		return
	}

	var remaining int
	if s.paused {
		// When paused, duration holds the remaining time
		remaining = s.duration
	} else {
		elapsed := int(time.Now().Unix() - s.startTime)
		remaining = s.duration - elapsed
	}

	if remaining <= 0 {
		fmt.Printf("%s session complete!\n", strings.Title(s.sessionType))
		return
	}

	statusLabel := ""
	if s.paused {
		statusLabel = " (paused)"
	}
	fmt.Printf("%s: %s remaining%s\n", strings.Title(s.sessionType), formatDuration(remaining), statusLabel)
}

func pause() {
	s, err := readState()
	if err != nil {
		fmt.Println("No timer running")
		return
	}

	if s.paused {
		fmt.Println("Timer already paused")
		return
	}

	// Calculate remaining time and store it
	elapsed := int(time.Now().Unix() - s.startTime)
	remaining := s.duration - elapsed

	if remaining <= 0 {
		fmt.Printf("%s session complete!\n", strings.Title(s.sessionType))
		return
	}

	// Store remaining time in duration field, mark as paused
	s.duration = remaining
	s.paused = true
	writeState(s)

	fmt.Printf("Paused with %s remaining\n", formatDuration(remaining))
}

func resume() {
	s, err := readState()
	if err != nil {
		fmt.Println("No timer running")
		return
	}

	if !s.paused {
		fmt.Println("Timer is not paused")
		return
	}

	// Start fresh with the remaining time
	s.startTime = time.Now().Unix()
	s.paused = false
	writeState(s)

	fmt.Printf("Resumed: %s remaining\n", formatDuration(s.duration))
}

func stop() {
	err := os.Remove(stateFile)
	if err != nil {
		fmt.Println("No timer running")
		return
	}
	fmt.Println("Timer stopped")
}

func watch() {
	s, err := readState()
	if err != nil {
		fmt.Println("No timer running")
		return
	}

	if s.paused {
		fmt.Println("Timer is paused - resume first")
		return
	}

	fmt.Printf("Watching %s session...\n", s.sessionType)

	for {
		s, err = readState()
		if err != nil {
			// Timer was stopped
			return
		}

		if s.paused {
			fmt.Println("Timer paused")
			return
		}

		elapsed := int(time.Now().Unix() - s.startTime)
		remaining := s.duration - elapsed

		if remaining <= 0 {
			title := strings.Title(s.sessionType) + " Complete!"
			message := "Time for a " + oppositeSession(s.sessionType)
			notify(title, message)
			fmt.Println(title)
			os.Remove(stateFile)
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func oppositeSession(sessionType string) string {
	if sessionType == "work" {
		return "break"
	}
	return "work session"
}

func readState() (state, error) {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return state{}, err
	}

	parts := strings.Split(string(data), "|")
	if len(parts) != 4 {
		return state{}, fmt.Errorf("invalid state format")
	}

	startTime, _ := strconv.ParseInt(parts[0], 10, 64)
	duration, _ := strconv.Atoi(parts[1])

	return state{
		startTime:   startTime,
		duration:    duration,
		sessionType: parts[2],
		paused:      parts[3] == "paused",
	}, nil
}

func writeState(s state) {
	statusStr := "running"
	if s.paused {
		statusStr = "paused"
	}
	data := fmt.Sprintf("%d|%d|%s|%s", s.startTime, s.duration, s.sessionType, statusStr)
	os.WriteFile(stateFile, []byte(data), 0644)
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func notify(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		// macOS: use osascript
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
		exec.Command("osascript", "-e", script).Run()

	case "linux":
		// Linux: use notify-send (common on most desktop environments)
		exec.Command("notify-send", title, message).Run()

	case "windows":
		// Windows: use PowerShell
		script := fmt.Sprintf(`[System.Reflection.Assembly]::LoadWithPartialName('System.Windows.Forms'); [System.Windows.Forms.MessageBox]::Show('%s','%s')`, message, title)
		exec.Command("powershell", "-Command", script).Run()
	}
}
