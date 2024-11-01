package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func writeLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("could not write line: %v", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("could not flush to file: %v", err)
	}

	return nil
}

func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error clearing terminal:", err)
	}
}

// TODO refactor out cmd in favor of bufio,
// use this function purely for getting file contents as slice strings. send back for processing away the marks
func getLines(filePath string, startMark string, endMark string) ([]string, error) {
	cmd := exec.Command("sed", "-n", fmt.Sprintf("/## %s/,/## %s/p", startMark, endMark), filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing sed command: %w", err)
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 2 {
		lines = lines[1 : len(lines)-2]
	}

	for _, line := range lines {
		line = strings.Trim(line, " ")
	}

	return lines, nil
}
func sliceStrsHas(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
