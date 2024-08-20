package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mritd/bubbles/common"

	"github.com/mritd/bubbles/selector"
)

type model struct {
	sl selector.Model
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// By default, the prompt component will not return a "tea.Quit"
	// message unless Ctrl+C is pressed.
	//
	// If there is no error in the input, the prompt component returns
	// a "common.DONE" message when the Enter key is pressed.
	switch msg {
	case common.DONE:
		return m, tea.Quit
	}

	_, cmd := m.sl.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.sl.View()
}

type GitBranch struct {
	Name string
}

func sanityCheck() (inRepo bool) {
	_, err := exec.Command("git", "rev-parse").Output()
	//above command errors if we are not in a git repo.
	return err == nil
}

func main() {
	if !sanityCheck() {
		os.Stderr.WriteString("must be in Git repo")
		os.Exit(1)
		return
	}
	for _, arg := range os.Args {
		fmt.Println(arg)
	}
	branches, err := gatherBranches()

	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
	data := buildBranchData(branches)
	generic := make([]interface{}, 0)
	for _, v := range data.branches {
		generic = append(generic, v)
	}
	m := &model{
		sl: selector.Model{
			Data:       generic,
			PerPage:    5,
			HeaderFunc: selector.DefaultHeaderFuncWithAppend(fmt.Sprintf("Current Branch: %s >", data.current.Name)),
			SelectedFunc: func(m selector.Model, obj interface{}, gdIndex int) string {
				t := obj.(GitBranch)
				return common.FontColor(fmt.Sprintf("[%d] %s", gdIndex+1, t.Name), selector.ColorSelected)
			},
			UnSelectedFunc: func(m selector.Model, obj interface{}, gdIndex int) string {
				t := obj.(GitBranch)
				return common.FontColor(fmt.Sprintf(" %d. %s", gdIndex+1, t.Name), selector.ColorUnSelected)
			},
			FinishedFunc: func(s interface{}) string {
				return s.(GitBranch).Name + "\n"
			},
			FooterFunc: func(m selector.Model, obj interface{}, gdIndex int) string {
				t := m.Selected().(GitBranch)
				return common.FontColor(fmt.Sprint(t.Name), selector.ColorFooter)
			},
		},
	}
	p := tea.NewProgram(m)
	err = p.Start()
	if err != nil {
		log.Fatal(err)
	}
	if !m.sl.Canceled() {
		branch := m.sl.Selected().(GitBranch)
		fmt.Printf("switching to branch '%s'", branch.Name)
		cmd := exec.Command("git", "switch", branch.Name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	} else {
		log.Println("user canceled...")
	}
}

// print branches

func gatherBranches() (branches []string, err error) {
	val, err := exec.Command("git", "branch").Output()
	if err != nil {
		return branches, err
	}

	branches = strings.Split(strings.TrimSpace(string(val)), "\n")
	return branches, err
}

type BranchData struct {
	current  GitBranch
	branches []GitBranch
}

func buildBranchData(_branches []string) BranchData {
	var current GitBranch
	branches := make([]GitBranch, 0)
	for _, b := range _branches {
		if len(b) < 1 {
			continue
		}
		if b[0] == '*' {

			b = strings.TrimPrefix(b, "* ")
			current = GitBranch{Name: b}
		} else {
			b = strings.TrimSpace(b)
			branches = append(branches, GitBranch{Name: b})
		}
	}

	data := BranchData{current: current, branches: branches}

	return data
}
