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

// Reset
const Reset = "\033[0m"

// Colors
const (
	Black         = "\033[30m"
	Red           = "\033[31m"
	Green         = "\033[32m"
	Yellow        = "\033[33m"
	Blue          = "\033[34m"
	Magenta       = "\033[35m"
	Cyan          = "\033[36m"
	White         = "\033[37m"
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

// Background Colors
const (
	BgBlack         = "\033[40m"
	BgRed           = "\033[41m"
	BgGreen         = "\033[42m"
	BgYellow        = "\033[43m"
	BgBlue          = "\033[44m"
	BgMagenta       = "\033[45m"
	BgCyan          = "\033[46m"
	BgWhite         = "\033[47m"
	BgBrightBlack   = "\033[100m"
	BgBrightRed     = "\033[101m"
	BgBrightGreen   = "\033[102m"
	BgBrightYellow  = "\033[103m"
	BgBrightBlue    = "\033[104m"
	BgBrightMagenta = "\033[105m"
	BgBrightCyan    = "\033[106m"
	BgBrightWhite   = "\033[107m"
)

// Effects
const (
	Bold          = "\033[1m"
	Dim           = "\033[2m"
	Italic        = "\033[3m"
	Underline     = "\033[4m"
	Blink         = "\033[5m"
	Invert        = "\033[7m"
	Hidden        = "\033[8m"
	StrikeThrough = "\033[9m"
)

var scriptDir string = "/usr/local/bin/jz_ro-build/"
var configPath string = os.Getenv("HOME") + "/.config/journal_zro/config.cfg"
var subcommands = []string{"'new'", "'find'", "'merge'"}
var config map[string]string = make(map[string]string)
var resultsList []Entry
var ignoreList []string
var mergeList []Entry

// Defaults
var SAVEDIR string = os.Getenv("HOME") + "/Documents/Journal_Zro/"
var MERGE_DIR string = SAVEDIR + ".merges/"
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
func createEntry() {
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
func findEntries(args []string, entries []Entry) {
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

	// Walk the directory or search previous results
	if entries != nil {
		for _, entry := range entries {
			if matchesTags(entry.Tags, searchTagSet, *inclusive) {
				resultsList = append(resultsList, entry)
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
						resultsList = append(resultsList, Entry{Path: path, Info: info, MergeOriginals: originalEntries, Tags: tags})
					}
				} else {
					if matchesTags(tags, searchTagSet, *inclusive) {
						resultsList = append(resultsList, Entry{Path: path, Info: info, MergeOriginals: nil, Tags: tags})
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
		for _, res := range resultsList {
			if res.MergeOriginals != nil {
				ignoreList = append(ignoreList, res.MergeOriginals...)
			}
		}
		for _, res := range resultsList {
			if !contains(ignoreList, res.Info.Name()) {
				filtered = append(filtered, res)
			}
		}
		resultsList = filtered
	} else {
		var filtered []Entry
		for _, res := range resultsList {
			if res.MergeOriginals == nil {
				filtered = append(filtered, res)
			}
		}
		resultsList = filtered
	}

	if *ascending && *descending {
		fmt.Println("Error: cannot sort by both asc and desc")
		os.Exit(1)
	}

	// Sort results by date if necessary
	if *ascending {
		sort.Slice(resultsList, func(i, j int) bool {
			return resultsList[i].Info.ModTime().Before(resultsList[j].Info.ModTime())
		})
	} else if *descending {
		sort.Slice(resultsList, func(i, j int) bool {
			return resultsList[i].Info.ModTime().After(resultsList[j].Info.ModTime())
		})
	}

	//Display Results
	if len(resultsList) < 1 {
		fmt.Println("No entries found with these parameters")
		os.Exit(0)
	} else {
		if *first {
			openNvim(resultsList[0].Path, false)
		} else {
			optionsPrompt("RESULTS", resultsList, searchTags, "")
			return
		}
	}

}
func optionsPrompt(title string, entriesList []Entry, searchTags []string, message string) {
	clearTerminal()
	fmt.Println(Green, "SEARCH TAGS = ", Reset, strings.Join(searchTags, ","))
	fmt.Println("")
	switch title {
	case "MERGE LIST":
		fmt.Println(Blue, "=======MERGE LIST===============================================================", Reset)
	case "RESULTS":
		fmt.Println(Blue, "=========RESULTS================================================================", Reset)
	default:
		fmt.Println(BgBrightGreen, BrightCyan, "=========+++++++================================================================", Reset)
	}
	fmt.Println("")
	displayEntries(entriesList)
	fmt.Println(BrightMagenta, "=========OPTIONS================================================================", Reset)
	//Prompt User
	if title == "RESULTS" {
		// E.g. r -i finance
		fmt.Print(Magenta, "[R]efine current search: ", Reset, "r -[opts] [tag]...\n")
		// E.g. n -a health
		fmt.Print(Magenta, "[N]ew search: ", Reset, "n -[opts] [tag]...\n")
		// E.g. a 1 4 12
		fmt.Print(Magenta, "[A]dd entry to merge list: ", Reset, "a [number]...\n")
		// E.g. w
		fmt.Print(Magenta, "[W]hole list to merge list: ", Reset, "w\n")
		// E.g. d 1 4 12
		fmt.Print(Magenta, "[D]elete entry permanently: ", Reset, "d [number]...\n")
	}
	if title == "MERGE LIST" {
		// E.g. m 2024
		fmt.Print(Magenta, "[M]erge entries from merge list to single entry: ", Reset, "m [name]...\n")
		// E.g. d 2 12 6
		fmt.Print(Magenta, "[D]elete entries from merge list: ", Reset, "d [number]...\n")
		fmt.Print(Magenta, "[B]ack to results: ", Reset, "b\n")
	} else {
		// E.g. v
		fmt.Print(Magenta, "[V]iew current merge list: ", Reset, "v\n")
		// E.g. q
		fmt.Print(Magenta, "[Q]uit: ", Reset, "q\n")
		// E.g. 31
		fmt.Print(Magenta, "[#] Number of the file to open: ", Reset, "[number]\n")
	}
	if message != "" {
		fmt.Println(Red, "==========INFO==================================================================", Reset)
		fmt.Println(Red, message, Reset)
	}
	fmt.Print("Your decision: ")

	var input string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input = scanner.Text()
	}

	inputArr := strings.Split(strings.Trim(input, " "), " ")
	newCmd := inputArr[0]
	newArgs := inputArr[1:]

	if title == "RESULTS" {
		switch strings.ToLower(newCmd) {
		case "r":
			if len(entriesList) > 4 {
				findEntries(newArgs, entriesList)
			} else {
				optionsPrompt("RESULTS", resultsList, searchTags, "Refinement is only available when there are 5 or more results.")
				return
			}
		case "n":
			findEntries(newArgs, nil)
		case "a":
			var tempList []string
			for _, arg := range newArgs {
				selectedNumber, err := strconv.Atoi(arg)
				if err != nil || selectedNumber < 1 || selectedNumber > len(entriesList) {
					fmt.Println("Invalid selection:"+arg, err)
				}
				tempList = append(tempList, arg)
				mergeList = append(mergeList, entriesList[selectedNumber-1])
			}
			msg := strings.Join(tempList, ", ") + " added to merge list"
			optionsPrompt("RESULTS", resultsList, searchTags, msg)
			return
		case "w":
			mergeList = append(mergeList, resultsList...)
			optionsPrompt("RESULTS", resultsList, searchTags, "Whole of results added to merge list")
			return
		case "d":
			// TODO add confirmation
			for _, arg := range newArgs {
				selectedNumber, err := strconv.Atoi(arg)
				if err != nil || selectedNumber < 1 || selectedNumber > len(entriesList) {
					fmt.Println("Invalid selection:"+arg, err)
				}
				os.Remove(entriesList[selectedNumber-1].Path)
			}
		case "v":
			if len(mergeList) > 0 {
				optionsPrompt("MERGE LIST", mergeList, searchTags, "")
				return
			} else {
				optionsPrompt("RESULTS", resultsList, searchTags, "Add something to your merge list first...")
				return
			}

		case "q":
			os.Exit(0)
		default:
			selectedNumber, err := strconv.Atoi(input)
			if err != nil || selectedNumber < 1 || selectedNumber > len(entriesList) {
				fmt.Println("Invalid selection. Please enter a valid option.")
				optionsPrompt("RESULTS", resultsList, searchTags, "")
				return
			}
			openNvim(entriesList[selectedNumber-1].Path, false)
		}
	} else if title == "MERGE LIST" {
		switch strings.ToLower(newCmd) {
		case "m":
			if len(mergeList) > 1 && newArgs[0] != "" {
				newMerge, err := makeMergeEntry(strings.Join(newArgs, " "))
				if err != nil {
					fmt.Println("Error merging entries", err)
					optionsPrompt("RESULTS", resultsList, searchTags, "Your merge list still lives!")
					return
				} else {
					fmt.Println("Merge Successful")
					openNvim(newMerge.Path, false)
					os.Exit(0)
				}
			} else {
				optionsPrompt("RESULTS", resultsList, searchTags, "Add at least two entries to your merge list first...")
				return
			}
		case "b":
			optionsPrompt("RESULTS", resultsList, searchTags, "")
			return
		case "d":
			if len(newArgs) > 0 {
				for _, arg := range newArgs {
					selectedNumber, err := strconv.Atoi(arg)
					if err != nil || selectedNumber < 1 || selectedNumber > len(mergeList) {
						optionsPrompt("MERGE LIST", mergeList, searchTags, "Invalid selection. Please enter a valid option.")
						return
					}

					mergeList = append(
						mergeList[:selectedNumber-1],
						mergeList[selectedNumber:]...)
				}
				optionsPrompt("MERGE LIST", mergeList, searchTags, "Merge list updated")
				return
			}
		case "q":
			os.Exit(0)
		default:
			selectedNumber, err := strconv.Atoi(input)
			if err != nil || selectedNumber < 1 || selectedNumber > len(entriesList) {
				fmt.Println("Invalid selection. Please enter a valid option.")
				optionsPrompt("RESULTS", resultsList, searchTags, "")
				return
			}
			// TODO make new window optional
			openNvim(entriesList[selectedNumber-1].Path, false)
		}
	}

}

// TODO Random reminder function to show a random entry to remind you of it

// TODO Add all feature to add all results to mergeList
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
func getDate(file string) (string, error) {
	cmd := exec.Command("head", "-n", "1", file)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	date := strings.Trim(string(output), " ")
	return date, nil
}
func displayEntries(entries []Entry) error {

	for i, entry := range entries {
		date, err := getDate(entry.Path)
		if err != nil {
			fmt.Println("Error getting date from entry", err)
		}
		fmt.Println(Bold, Blue, strconv.Itoa(i+1)+") ", Reset, entry.Info.Name(), " | Created: ", date)
		preview, err := getLines(entry.Path, "Entry_", "_Entry")
		if err != nil {
			fmt.Println("Error reading body of entry at "+entry.Path, err)
			return err
		}
		if len(preview) < 1 {
			fmt.Println("No text available for preview")
		} else {
			for i, pLine := range preview {
				if i < 5 {
					fmt.Println("\t", Green+pLine, Reset)
				}
			}
		}
		//Separator
		if i < len(entries)-1 {

			fmt.Println(Yellow, "================================================================================", Reset)
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
func mergeEntries(list []Entry) {
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
		fmt.Println("Unknown command. Use 'new', 'find', or 'merge'.")
		os.Exit(1)
	}
}
