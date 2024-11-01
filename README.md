
# JournalZ-ro

JournalZ-ro is a command-line tool for creating, finding, and merging entries in your journal, designed for efficient and tag-based entry management.

## Features
- **New Entry**: Create a new entry with predefined templates.
- **Find Entries by Tag**: Quickly search through entries using tags.
- **Merge Entries**: Merge entries with matching tags into a single entry.

## Prerequisites
- **Neovim** and **Alacritty** are currently required, although future versions aim to reduce dependency requirements for greater compatibility.

## Installation

1. **Add JournalZ-ro to your PATH**  
   Add the `jz_ro-build` folder to your system's `$PATH`. Usually via your shell config file.
    e.g. ~/.zshrc for zsh, ~/.bashrc for bash

2. **Symlink to `/usr/local/bin`**  
   Create a symlink for easy access:
   ```bash
   sudo ln -s /path/to/jz_ro-build /usr/local/bin/
   ```

## Usage

### New Entry
Create a new journal entry:
```bash
journalz_ro new
```
This command generates a new entry based on the template defined in `entry_template` and opens it in a new neovim instance.

### Find Entries by Tag
Find entries associated with a specific tag:
```bash
journalz_ro find <tag>
```
Find entries then refine your search, start a new search, delete entries or add them to a merge list.

### Merge Entries
Merge entries that share a specific tag. This command requires a name for the merged entry:
```bash
journalz_ro merge <tag> <name>
```
The merged entry will be saved as `<name>` in the MERGE_DIR directory.

## Configuration

JournalZ-ro requires two configuration files in the `jz_ro-build` directory:
- `default.cfg`: Defines general settings.
- `entry_template`: Specifies the structure of a new entry.

Both files should be in the same folder as the executable for the app to function.

## Planned Features
1. Configuration File
    - .cfg file for specifying save paths and custom templates etc
2. Settings for Editor 
    - Not just neovim
    - Open in new window or current window

## Thanks

Contributions to JournalZ-ro are welcome! Please feel free to open issues or submit pull requests.

This is my first open-source tool and my first public build. I'm trying to learn more of everything, including git/github. Code reviews and opinions are welcome. 

Thanks again, and I hope this tool can help some people to learn smoothly. 

I made this tool because I wanted to take notes in my configured environment without having to switch brains or think much. Just bind `journalz_ro new` to a key combo, type my note, tag it and move on. No title thinking, just quick notes that can later be merged by tags. Then those can be taken and refined later. I think this approach has a lot of potential. 

---

Let me know if you need any further changes!
