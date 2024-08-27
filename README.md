# Git Switcher

A simple command line tool to switch between git branches, written in Go. It is leveraging the [bubbletea](https://github.com/charmbracelet/bubbletea) library to create a TUI.

### To try out the project

- clone this repo
- run `make thebuild`
- this will create a build directory and a binary file named `gitSwitcher`

Then, you can add an alias to this binary file in your `.bashrc` or `.zshrc` file.

```bash
alias gitme="path/to/gitSwitcher/binary"
```

Once this is added, navigate to a git repository and try it out

```
gitme
```
