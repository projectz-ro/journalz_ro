package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var scriptDir string = "/usr/local/bin/journal_zro"
var configPath string = os.Getenv("HOME") + "/.config/journal_zro/config.cfg"
var subcommands = []string{"'new'", "'find'", "'merge'"}
var config map[string]string = make(map[string]string)
var SAVEDIR string = os.Getenv("HOME") + "/Documents/Journal_Zro/"
var TEMPLATE string = scriptDir + "/entry_template.md"

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

func findNotes(tags []string) {
	fmt.Printf("Searching for notes with tags: %s\n", strings.Join(tags, ", "))
	// Example: Search logic for tags in files under SAVEDIR
	// Skipping implementation for simplicity
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
		findCmd := flag.NewFlagSet("find", flag.ExitOnError)
		findCmd.Parse(os.Args[2:])
		if len(findCmd.Args()) == 0 {
			fmt.Println("Error: You must provide at least one tag to find.")
			os.Exit(1)
		}
		findNotes(findCmd.Args())

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
