package commands

import ()

func createEntry() {
	currentDate := time.Now().Format("01/02/2006")

	entryCount, err := countEntries()
	if err != nil {
		fmt.Println("Error counting entries")
		return
	}

	title := "Entry" + strconv.Itoa(entryCount) + ".md"
	filepath := SAVE_DIR + "/" + title
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
