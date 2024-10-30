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

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

var scriptDir string = "/usr/local/bin/jz_ro-build/"
var configPath string = os.Getenv("HOME") + "/.config/journal_zro/config.cfg"
var subcommands = []string{"'new'", "'find'", "'merge'"}
var config map[string]string = make(map[string]string)

// Defaults
var SAVEDIR string = os.Getenv("HOME") + "/Documents/Journal_Zro/"
var MERGE_DIR string = SAVEDIR + "/.merges/"
var TEMPLATE string = scriptDir + "/entry_template.md"

type Entry struct {
	Path           string
	Info           os.FileInfo
	MergeOriginals []string
	Tags           []string
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}
	return err == nil
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
	files, err := os.ReadDir(SAVEDIR)
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
func createNote() {
	currentDate := time.Now().Format("01/02/2006")

	entryCount, err := countEntries()
	if err != nil {
		fmt.Println("Error counting entries")
		return
	}

	title := "Entry" + strconv.Itoa(entryCount) + ".md"
	filepath := SAVEDIR + "/" + title

	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	newNote, err := os.ReadFile(TEMPLATE)
	if err != nil {
		fmt.Println("Error reading template file:", err)
		return
	}

	re := regexp.MustCompile(`MM/DD/YYYY`)
	newNote = re.ReplaceAll(newNote, []byte(currentDate))

	err = os.WriteFile(filepath, newNote, 0644)
	if err != nil {
		fmt.Println("Error copying default.cfg to config file. ", err)
		return
	}

	openNvim(filepath, true)

}
func findNotes(args []string, entries []Entry) {
	findCmd := flag.NewFlagSet("find", flag.ExitOnError)

	// Flags
	inclusive := findCmd.Bool("i", false, "Inclusive search: show entries which include ANY of the provided tags (default: all tags must match)")
	first := findCmd.Bool("f", false, "Return only the first file to match the provided tags")
	ascending := findCmd.Bool("a", false, "Sort by date/time in ascending order")
	descending := findCmd.Bool("d", false, "Sort by date/time in descending order")
	originalsOnly := findCmd.Bool("o", false, "Originals only, do not include merged entries in the results (default: prioritize merge entries and ignore originals if they're contained in a merge)")

	findCmd.Parse(args)

	searchTags := findCmd.Args()
	if len(searchTags) == 0 {
		fmt.Println("Error: You must provide at least one tag to find.")
		os.Exit(1)
	}

	//Make map for comparison later
	searchTagSet := make(map[string]bool)
	for i := range searchTags {
		searchTags[i] = strings.ToLower(strings.Trim(searchTags[i], " "))
		searchTagSet[searchTags[i]] = true
	}

	var results []Entry
	var ignore []string

	// Walk the directory or search previous results
	if entries != nil {
		for _, entry := range entries {
			if matchesTags(entry.Tags, searchTagSet, *inclusive) {
				results = append(results, entry)
			}
		}

	} else {

		err := filepath.Walk(SAVEDIR, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				tags, err := getLines(path, "Tags_", "_Tags")
				if err != nil {
					fmt.Println("Error fetching tags from file:", err)
					return err
				}
				if strings.Contains(path, MERGE_DIR) {
					if !*originalsOnly {
						originalEntries, err := getLines(path, "Originals_", "_Originals")
						if err != nil {
							fmt.Println("Error fetching Originals from file:", err)
							return err
						}
						results = append(results, Entry{Path: path, Info: info, MergeOriginals: originalEntries, Tags: tags})
					}
				} else {
					if matchesTags(tags, searchTagSet, *inclusive) {
						results = append(results, Entry{Path: path, Info: info, MergeOriginals: nil, Tags: tags})
					}
				}
			}
			return nil
		})

		if err != nil && err.Error() != "stop early" {
			fmt.Println("Error walking file tree:", err)
			os.Exit(1)
		}
	}
	// Default
	if !*originalsOnly {
		var filtered []Entry
		for _, res := range results {
			if res.MergeOriginals != nil {
				ignore = append(ignore, res.MergeOriginals...)
			}
		}
		for _, res := range results {
			if !contains(ignore, res.Info.Name()) {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	} else {
		var filtered []Entry
		for _, res := range results {
			if res.MergeOriginals == nil {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	if *ascending && *descending {
		fmt.Println("Error: cannot sort by both asc and desc")
		os.Exit(1)
	}

	// Sort results by date if necessary
	if *ascending {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Info.ModTime().Before(results[j].Info.ModTime())
		})
	} else if *descending {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Info.ModTime().After(results[j].Info.ModTime())
		})
	}

	//Display Results
	if len(results) < 1 {
		fmt.Println("No entries found with these parameters")
		os.Exit(0)
	} else {
		if *first {
			openNvim(results[0].Path, false)
		} else {
			resultsPrompt(results, searchTags)
		}
	}

}
func resultsPrompt(results []Entry, searchTags []string) {
	clearTerminal()
	fmt.Println(green, "SEARCH TAGS = ", reset, strings.Join(searchTags, ","))
	outputResults(results)
	fmt.Println(red, "=========OPTIONS==========================================================================", reset)
	//Prompt User
	fmt.Print("[R]efine current search: r -[opts] [tag]...\n")
	fmt.Print("or\n")
	fmt.Print("[N]ew search: n -[opts] [tag]...\n")
	fmt.Print("or\n")
	fmt.Print("[Q]uit\n")
	fmt.Print("or\n")
	fmt.Print("Enter the number of the file you want to open: ")

	var input string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = scanner.Text()
	}

	secondCmd := strings.Split(strings.Trim(input, " "), " ")

	if strings.ToLower(secondCmd[0]) == "r" {
		findNotes(secondCmd[1:], results)
	} else if strings.ToLower(secondCmd[0]) == "n" {
		findNotes(secondCmd[1:], nil)
	} else if strings.ToLower(secondCmd[0]) == "q" {
		os.Exit(0)
	} else {
		selectedNumber, err := strconv.Atoi(input)
		if err != nil || selectedNumber < 1 || selectedNumber > len(results) {
			fmt.Println("Invalid selection. Please enter a valid option.")
			resultsPrompt(results, searchTags)
			return
		}
		openNvim(results[selectedNumber-1].Path, false)
	}

}
func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error clearing terminal:", err)
	}
}
func getDate(file string) (string, error) {
	cmd := exec.Command("head", "-n", "1", file)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	date := strings.Trim(string(output), " ")
	return date, nil
}
func outputResults(results []Entry) error {

	fmt.Println(blue, "=========RESULTS==========================================================================", reset)
	for i, res := range results {
		date, err := getDate(res.Path)
		if err != nil {
			fmt.Println("Error getting date from entry", err)
		}
		fmt.Println(blue, strconv.Itoa(i+1)+") ", reset, res.Info.Name(), " | Created: ", date)
		preview, err := getLines(res.Path, "Entry_", "_Entry")
		if err != nil {
			fmt.Println("Error reading body of entry at "+res.Path, err)
			return err
		}
		if len(preview) < 1 {
			fmt.Println("No text available for preview")
		} else {
			for i, pLine := range preview {
				if i < 5 {
					fmt.Println("\t", green+pLine, reset)
				}
			}
		}
		//Separator
		if i < len(results)-1 {

			fmt.Println(yellow, "==========================================================================================", reset)
		}
	}
	return nil
}
func contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
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
func getLines(file string, startMark string, endMark string) ([]string, error) {
	cmd := exec.Command("sed", "-n", fmt.Sprintf("/## %s/,/## %s/p", startMark, endMark), file)

	// Execute the command and capture the output
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing sed command: %w", err)
	}
	// fmt.Println("SED: ", output)

	// Convert the output to a string and split by newlines
	lines := strings.Split(string(output), "\n")

	// Remove the first and last lines which are the markers
	if len(lines) > 2 {
		lines = lines[1 : len(lines)-2]
	}

	for _, line := range lines {
		line = strings.Trim(line, " ")
	}

	return lines, nil
}
func mergeNotes(name string, tags []string) {
	fmt.Printf("Merging notes with tags: %s into %s\n", strings.Join(tags, ", "), name)
	// Example: Merge logic for notes that match the tags
	// Skipping implementation for simplicity
}
func main() {

	// Set config if exists; else create it
	config, err := loadConfig(scriptDir)
	if err != nil {
		fmt.Println("Error loading config file: ", err)
		return
	}

	if config["SAVE_DIR"] != "" {
		SAVEDIR = os.Getenv("HOME") + "/" + config["SAVE_DIR"]
	}
	if _, err := os.Stat(SAVEDIR); os.IsNotExist(err) {
		err := os.MkdirAll(SAVEDIR, 0755)
		if err != nil {
			fmt.Println("Error creating save directory", err)
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
		createNote()

	case "find":
		if len(os.Args) > 2 {
			findNotes(os.Args[2:], nil)
		} else {
			fmt.Println("Error: You must provide at least one argument")
		}

	case "merge":
		mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
		mergeCmd.Parse(os.Args[2:])
		if len(mergeCmd.Args()) < 2 {
			fmt.Println("Error: You must provide a name and at least one tag to merge.")
			os.Exit(1)
		}
		name := mergeCmd.Args()[0]
		tags := mergeCmd.Args()[1:]
		mergeNotes(name, tags)

	default:
		fmt.Println("Unknown command. Use 'new', 'find', or 'merge'.")
		os.Exit(1)
	}
}
