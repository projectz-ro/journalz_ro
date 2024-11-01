package commands

import ()

func findEntries(args []string, entries []Entry) {
	// TODO parse flags yourself, like in secondCmd, that way I can do combined flags
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

		err := filepath.Walk(SAVE_DIR, func(path string, info os.FileInfo, err error) error {
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
			if !sliceStrsHas(ignoreList, res.Info.Name()) {
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
