package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const stateFile = "/tmp/pomogori.state"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pomogori [start|status]")
		return
	}

	switch os.Args[1] {
	case "start":
		now := time.Now().Unix()
		os.WriteFile(stateFile, []byte(fmt.Sprintf("%d", now)), 0644)
		fmt.Println("Timer started")

	case "status":
		data, err := os.ReadFile(stateFile)
		if err != nil {
			fmt.Println("No timer running")
			return
		}

		startTime, _ := strconv.ParseInt(string(data), 10, 64)
		elapsed := time.Now().Unix() - startTime
		fmt.Printf("Running: %d seconds elapsed\n", elapsed)
	}
}
