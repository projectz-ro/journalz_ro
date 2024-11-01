package ui

import ()

// Colors
const (
	Reset         = "\033[0m"
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

	// Background Colors
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

	// Effects
	Bold          = "\033[1m"
	Dim           = "\033[2m"
	Italic        = "\033[3m"
	Underline     = "\033[4m"
	Blink         = "\033[5m"
	Invert        = "\033[7m"
	Hidden        = "\033[8m"
	StrikeThrough = "\033[9m"
)

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

// TODO break this up into logic vs ui
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
