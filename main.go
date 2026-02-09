package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const version = "1.0.0"

var (
	dataDir     string
	stateFile   string
	historyFile string
	configFile  string
	cfg         config
)

type config struct {
	WorkMinutes  int
	BreakMinutes int
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	dataDir = filepath.Join(home, ".pomogori")
	stateFile = filepath.Join(dataDir, "state")
	historyFile = filepath.Join(dataDir, "history")
	configFile = filepath.Join(dataDir, "config")

	// Ensure data directory exists
	os.MkdirAll(dataDir, 0755)

	// Load config with defaults
	cfg = config{
		WorkMinutes:  25,
		BreakMinutes: 5,
	}
	loadConfig()
}

func loadConfig() {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return // Use defaults
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "work_minutes":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.WorkMinutes = v
			}
		case "break_minutes":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.BreakMinutes = v
			}
		}
	}
}

type state struct {
	startTime   int64
	duration    int
	sessionType string
	paused      bool
}

type historyEntry struct {
	timestamp   int64
	duration    int
	sessionType string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "work", "w":
		startCmd := flag.NewFlagSet("work", flag.ExitOnError)
		duration := startCmd.Int("d", cfg.WorkMinutes, "duration in minutes")
		startCmd.Parse(os.Args[2:])
		start("work", *duration*60)

	case "break", "b":
		breakCmd := flag.NewFlagSet("break", flag.ExitOnError)
		duration := breakCmd.Int("d", cfg.BreakMinutes, "duration in minutes")
		breakCmd.Parse(os.Args[2:])
		start("break", *duration*60)

	case "status", "s":
		status()

	case "pause", "p":
		pause()

	case "resume", "r":
		resume()

	case "watch":
		watch()

	case "stop":
		stop()

	case "stats":
		stats()

	case "config":
		showConfig()

	case "help", "h", "-h", "--help":
		printHelp()

	case "version", "v", "-v", "--version":
		fmt.Printf("pomogori %s\n", version)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Run 'pomogori help' for usage")
	}
}

func printUsage() {
	fmt.Println("pomogori - A simple pomodoro timer for the terminal")
	fmt.Println()
	fmt.Println("Usage: pomogori <command> [options]")
	fmt.Println()
	fmt.Println("Run 'pomogori help' for detailed usage")
}

func printHelp() {
	help := `pomogori - A simple pomodoro timer for the terminal

USAGE
    pomogori <command> [options]

COMMANDS
    work, w      Start a work session (default: %d minutes)
    break, b     Start a break session (default: %d minutes)
    status, s    Show current timer status
    pause, p     Pause the current timer
    resume, r    Resume a paused timer
    watch        Watch timer with live countdown display
    stop         Stop and discard the current timer
    stats        Show session statistics
    config       Show configuration and file paths
    help, h      Show this help message
    version, v   Show version

OPTIONS
    -d <minutes>   Set custom duration for work/break

EXAMPLES
    pomogori work              Start a 25-minute work session
    pomogori w -d 50           Start a 50-minute work session
    pomogori break             Start a 5-minute break
    pomogori b -d 15           Start a 15-minute break
    pomogori watch             Watch the timer countdown
    pomogori status            Check remaining time

WORKFLOW
    1. pomogori work           Start working
    2. pomogori watch          Watch the countdown (optional)
    3. [notification]          Get notified when done
    4. pomogori break          Take a break
    5. Repeat!

CONFIG
    Edit ~/.pomogori/config to change defaults:
        work_minutes = 25
        break_minutes = 5

VERSION
    %s
`
	fmt.Printf(help, cfg.WorkMinutes, cfg.BreakMinutes, version)
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

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print("\033[?25h") // Show cursor
		fmt.Println("\nInterrupted")
		os.Exit(0)
	}()

	// Hide cursor for cleaner display
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	for {
		s, err = readState()
		if err != nil {
			// Timer was stopped
			fmt.Println("\nTimer stopped")
			return
		}

		if s.paused {
			fmt.Println("\nTimer paused")
			return
		}

		elapsed := int(time.Now().Unix() - s.startTime)
		remaining := s.duration - elapsed

		if remaining <= 0 {
			clearLine()
			title := strings.Title(s.sessionType) + " Complete!"
			message := "Time for a " + oppositeSession(s.sessionType)
			recordSession(s.sessionType, s.duration)
			notify(title, message)
			fmt.Println(title)
			os.Remove(stateFile)
			return
		}

		// Display live countdown
		clearLine()
		progress := float64(elapsed) / float64(s.duration)
		bar := renderProgressBar(progress, 20)
		fmt.Printf("\r  %s  %s  %s",
			strings.Title(s.sessionType),
			formatDuration(remaining),
			bar)

		time.Sleep(1 * time.Second)
	}
}

func clearLine() {
	fmt.Print("\r\033[K") // Carriage return + clear to end of line
}

func renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("=", filled)
	empty := strings.Repeat("-", width-filled)

	return fmt.Sprintf("[%s%s]", bar, empty)
}

func oppositeSession(sessionType string) string {
	if sessionType == "work" {
		return "break"
	}
	return "work session"
}

func recordSession(sessionType string, duration int) {
	// Append to history file: timestamp|duration|sessionType
	entry := fmt.Sprintf("%d|%d|%s\n", time.Now().Unix(), duration, sessionType)

	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(entry)
}

func readHistory() []historyEntry {
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil
	}

	var entries []historyEntry
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}

		ts, _ := strconv.ParseInt(parts[0], 10, 64)
		dur, _ := strconv.Atoi(parts[1])
		entries = append(entries, historyEntry{
			timestamp:   ts,
			duration:    dur,
			sessionType: parts[2],
		})
	}

	return entries
}

func showConfig() {
	fmt.Println("=== Pomogori Config ===")
	fmt.Println()
	fmt.Printf("Config file: %s\n", configFile)
	fmt.Printf("Data dir:    %s\n", dataDir)
	fmt.Println()
	fmt.Println("Current settings:")
	fmt.Printf("  work_minutes  = %d\n", cfg.WorkMinutes)
	fmt.Printf("  break_minutes = %d\n", cfg.BreakMinutes)
	fmt.Println()

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("No config file found. Create one with:")
		fmt.Printf("  echo 'work_minutes = 25' >> %s\n", configFile)
		fmt.Printf("  echo 'break_minutes = 5' >> %s\n", configFile)
	}
}

func stats() {
	entries := readHistory()
	if len(entries) == 0 {
		fmt.Println("No completed sessions yet")
		return
	}

	// Calculate stats
	var totalWork, totalBreak int
	var workCount, breakCount int
	var todayWork, todayBreak int

	today := time.Now().Truncate(24 * time.Hour)

	for _, e := range entries {
		if e.sessionType == "work" {
			totalWork += e.duration
			workCount++
			if time.Unix(e.timestamp, 0).After(today) {
				todayWork += e.duration
			}
		} else {
			totalBreak += e.duration
			breakCount++
			if time.Unix(e.timestamp, 0).After(today) {
				todayBreak += e.duration
			}
		}
	}

	fmt.Println("=== Pomogori Stats ===")
	fmt.Println()
	fmt.Println("Today:")
	fmt.Printf("  Work:  %s (%d sessions)\n", formatDurationLong(todayWork), countToday(entries, "work"))
	fmt.Printf("  Break: %s\n", formatDurationLong(todayBreak))
	fmt.Println()
	fmt.Println("All Time:")
	fmt.Printf("  Work:  %s (%d sessions)\n", formatDurationLong(totalWork), workCount)
	fmt.Printf("  Break: %s (%d sessions)\n", formatDurationLong(totalBreak), breakCount)
}

func countToday(entries []historyEntry, sessionType string) int {
	today := time.Now().Truncate(24 * time.Hour)
	count := 0
	for _, e := range entries {
		if e.sessionType == sessionType && time.Unix(e.timestamp, 0).After(today) {
			count++
		}
	}
	return count
}

func formatDurationLong(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
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
