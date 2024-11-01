package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	scriptDir   string            = "/usr/local/bin/jz_ro-build/"
	configPath  string            = os.Getenv("HOME") + "/.config/journal_zro/config.cfg"
	subcommands []string          = []string{"'new'", "'find'", "'merge'"}
	config      map[string]string = make(map[string]string)
	resultsList []Entry
	ignoreList  []string
	mergeList   []Entry
	SAVE_DIR    string = os.Getenv("HOME") + "/Documents/Journal_Zro/"
	MERGE_DIR   string = SAVE_DIR + ".merges/"
	TEMPLATE    string = scriptDir + "/entry_template.md"
)

type Entry struct {
	Path           string
	Info           os.FileInfo
	MergeOriginals []string
	Tags           []string
}

func loadConfig(scriptPath string) (map[string]string, error) {
	configDir := os.Getenv("HOME") + "/.config/journal_zro/"

	if _, err := os.Stat(configDir); os.IsNotExist(err) {

		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			fmt.Println("Error creating config directory", err)
			return nil, err
		}
	}

	if !fileExists(configPath) {
		newConf, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		defer newConf.Close()

		defaultConfig, err := os.OpenFile(scriptPath+"/default.cfg", os.O_RDWR, 0755)
		if err != nil {
			return nil, err
		}
		defer defaultConfig.Close()

		_, errCopy := io.Copy(newConf, defaultConfig)
		if errCopy != nil {
			fmt.Println("Error copying default.cfg to config file. ", err)
			return nil, err
		}
	}

	file, err := os.OpenFile(configPath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Split the line into key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config line: %s", line)
		}

		// Trim spaces and store key-value pair
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
func countEntries() (int, error) {
	files, err := os.ReadDir(SAVE_DIR)
	if err != nil {
		return 0, err
	}

	mdRegex := regexp.MustCompile(`\.md$`)
	mdCount := 0
	for _, file := range files {
		if !file.IsDir() && mdRegex.MatchString(file.Name()) {
			mdCount++
		}
	}
	return mdCount, nil
}
func openNvim(filePath string, insertMode bool) {
	cmdArgs := []string{"-e", "nvim", "+" + config["START_POS"], filePath}

	if insertMode {
		cmdArgs = append(cmdArgs, "-c", "startinsert")
	}
	cmd := exec.Command(config["TERMINAL_APP"], cmdArgs...)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error opening terminal or neovim", err)
		return
	}
}

// TODO Random reminder function to show a random entry to remind you of it

func makeMergeEntry(name string) (Entry, error) {
	var newMerge Entry
	newMerge.Path = MERGE_DIR + "/" + name + ".md"
	var entryLines []string
	var allLines []string
	for _, entry := range mergeList {
		newMerge.Tags = append(newMerge.Tags, entry.Tags...)
		//if merge file, add its list of originals
		if len(entry.MergeOriginals) > 0 {
			newMerge.MergeOriginals = append(newMerge.MergeOriginals, entry.MergeOriginals...)
		}
		lines, err := getLines(entry.Path, "Entry_", "_Entry")
		if err != nil {
			fmt.Println("Error reading lines from file", err)
			return newMerge, err
		}
		newMerge.MergeOriginals = append(newMerge.MergeOriginals, entry.Info.Name())
		entryLines = append(entryLines, lines...)
	}
	// Write Merge
	allLines = append(allLines, "                                                                      "+time.Now().Format("01/02/2006"))
	allLines = append(allLines, "---")
	allLines = append(allLines, "## Entry_")
	allLines = append(allLines, entryLines...)
	allLines = append(allLines, "## _Entry")
	allLines = append(allLines, "---")
	allLines = append(allLines, "")
	allLines = append(allLines, "## Tags_")
	allLines = append(allLines, newMerge.Tags...)
	allLines = append(allLines, "## _Tags")
	allLines = append(allLines, "")
	allLines = append(allLines, "## Originals_")
	allLines = append(allLines, newMerge.MergeOriginals...)
	allLines = append(allLines, "## _Originals")

	err := writeLines(newMerge.Path, allLines)
	if err != nil {
		fmt.Println("Error writing merge file", err)
		os.Exit(1)
	}
	return newMerge, nil
}

// TODO Should be removed in favor of getLines with line count arg for getting only first line
func getDate(file string) (string, error) {
	cmd := exec.Command("head", "-n", "1", file)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	date := strings.Trim(string(output), " ")
	return date, nil
}

func matchesTags(tags []string, searchTagSet map[string]bool, inclusive bool) bool {
	if inclusive {
		// Inclusive search: check if any tag matches
		for _, tag := range tags {
			if searchTagSet[strings.ToLower(tag)] {
				return true
			}
		}
		return false
	} else {
		// Exclusive search: all tags must match
		for sTag := range searchTagSet {
			found := false
			for _, tag := range tags {
				if sTag == strings.ToLower(tag) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
}

func main() {

	// Set config if exists; else create it
	config, err := loadConfig(scriptDir)
	if err != nil {
		fmt.Println("Error loading config file: ", err)
		return
	}

	if config["SAVE_DIR"] != "" {
		SAVE_DIR = os.Getenv("HOME") + "/" + config["SAVE_DIR"]
	}
	if _, err := os.Stat(SAVE_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(SAVE_DIR, 0755)
		if err != nil {
			fmt.Println("Error creating save directory", err)
			return
		}
	}
	if _, err := os.Stat(MERGE_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(MERGE_DIR, 0755)
		if err != nil {
			fmt.Println("Error creating merge directory", err)
			return
		}
	}
	if len(os.Args) < 2 {
		fmt.Println("Expected " + strings.Join(subcommands, ", ") + "subcommands.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		newCmd := flag.NewFlagSet("new", flag.ExitOnError)
		newCmd.Parse(os.Args[2:])
		createEntry()
	case "find":
		if len(os.Args) > 2 {
			findEntries(os.Args[2:], nil)
		} else {
			fmt.Println("Error: You must provide at least one argument")
		}
	default:
		fmt.Println("Unknown command. Use 'new', 'find'.")
		os.Exit(1)
	}
}
