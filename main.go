package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mritd/bubbles/common"

	"github.com/mritd/bubbles/selector"
)

type model struct {
	insertMode     bool
	unfilteredData []GitBranch
	ti             textinput.Model
	sl             selector.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func fakeKeyMsg(kt tea.KeyType, r rune) tea.KeyMsg {
	return tea.KeyMsg{
		Type:  kt,
		Runes: []rune{r},
		Alt:   false,
		Paste: false,
	}
}

const shouldlog = false

func tlog(str string) {
	if !shouldlog {
		return
	}
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	f.WriteString(str)
	f.WriteString("\n")
	defer f.Close()
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

	if m.insertMode {
		return m.handleInsertMode(msg)
	}
	tlog("in Update")
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "i":
			tlog("pressed i")
			m.insertMode = true
			m.ti.Focus()
			return m, nil
		case "j":
			fm := fakeKeyMsg(tea.KeyDown, 'j')
			_, cmd := m.sl.Update(fm)
			return m, cmd
		case "k":
			fm := fakeKeyMsg(tea.KeyUp, 'k')
			_, cmd := m.sl.Update(fm)
			return m, cmd
		}
	}
	_, cmd := m.sl.Update(msg)
	return m, cmd
}

func (m *model) updateData() selector.Model {
	// make a new slice of strings
	data := make([]interface{}, 0)
	// filter the git branches based on filter
	for _, branch := range m.unfilteredData {
		if fuzzy.Match(m.ti.Value(), branch.Name) {
			data = append(data, branch)
			tlog(fmt.Sprintf("it's a match! %s - %s", m.ti.Value(), branch.Name))
		}
	}
	// set selector data to new slice.

	return newSelector(data)
}

func (m *model) handleInsertMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "enter":
			_, cmd := m.sl.Update(msg)
			return m, cmd
		case "esc":
			m.insertMode = false
			m.ti.Blur()
			return m, nil
		case "ctrl+c":
			m.insertMode = false
			return m, tea.Quit
		default:

			ti, cmd := m.ti.Update(msg)
			m.ti = ti
			sl := m.updateData()
			m.sl = sl
			_, slcmd := m.sl.Update(msg)
			return m, tea.Batch(cmd, slcmd)
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(m.ti.View())
	b.WriteString("\n")
	b.WriteString(m.sl.View())
	return b.String()
}

type GitBranch struct {
	Name string
}

func sanityCheck() (inRepo bool) {
	_, err := exec.Command("git", "rev-parse").Output()
	//above command errors if we are not in a git repo.
	return err == nil
}
func newSelector(data []interface{}) selector.Model {
	return selector.Model{
		Data: data,
		// PerPage:    5,
		HeaderFunc: func(m selector.Model, obj interface{}, gdIndex int) string { return "" },
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
		}}
}

func main() {
	if !sanityCheck() {
		os.Stderr.WriteString("must be in Git repo")
		os.Exit(1)
		return
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
	ti := textinput.New()
	m := &model{
		insertMode:     false,
		unfilteredData: data.branches,
		ti:             ti,
		sl:             newSelector(generic),
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
