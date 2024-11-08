package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mritd/bubbles/common"
	"github.com/mritd/bubbles/selector"
)

type stashUi struct {
	canceled      bool
	changes       []string
	stagedChanges []string
	sl            selector.Model
}

func (m *stashUi) Init() tea.Cmd {
	return nil
}
func (m *stashUi) View() string {
	var builder strings.Builder

	builder.WriteString(m.sl.View())
	builder.WriteString("\n\n")
	for i, v := range m.stagedChanges {
		str := common.FontColor(fmt.Sprintf(" %d. %s", i+1, v), selector.ColorUnSelected)
		builder.WriteString(str)
		builder.WriteString("\n")
	}
	return builder.String()
}

func (m *stashUi) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg == common.DONE {
		return m, tea.Quit
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch strings.ToLower(msg.String()) {
		case "ctrl+c", "q":
			m.canceled = true
			return m, tea.Quit
		case "s":
			return m.stageChange(msg)
		case "u":
			return m.undoStage(msg)
		}
	}
	_, cmd := m.sl.Update(msg)
	return m, cmd
}

func (m *stashUi) undoStage(msg tea.Msg) (tea.Model, tea.Cmd) {
	newChanges := make([]string, len(m.changes))
	copy(newChanges, m.changes)
	newStaged := make([]string, 0)
	for i, v := range m.stagedChanges {
		if i == 0 {
			newChanges = append(newChanges, v)
		} else {
			newStaged = append(newStaged, v)
		}
	}
	toAny := make([]any, 0)
	for _, val := range newChanges {
		toAny = append(toAny, val)
	}
	newSl := makeSelector(toAny)
	m.changes = newChanges
	m.stagedChanges = newStaged
	m.sl = newSl
	_, cmd := m.sl.Update(msg)
	return m, cmd
}

func (m *stashUi) stageChange(msg tea.Msg) (tea.Model, tea.Cmd) {
	idx := m.sl.Index()
	newChanges := make([]string, 0)
	newStaged := make([]string, len(m.stagedChanges))
	copy(newStaged, m.stagedChanges)
	for i, val := range m.changes {
		if i == idx {
			tlog(fmt.Sprintf("adding %v to staged", val))
			newStaged = append(newStaged, val)
		} else {
			tlog(fmt.Sprintf("adding %v to changes", val))
			newChanges = append(newChanges, val)
		}
	}
	toAny := make([]any, 0)
	for _, val := range newChanges {
		toAny = append(toAny, val)
	}
	newSl := makeSelector(toAny)
	m.changes = newChanges
	m.stagedChanges = newStaged
	m.sl = newSl
	_, cmd := m.sl.Update(msg)
	return m, cmd
}

func makeSelector(data []any) selector.Model {
	return selector.Model{
		Data:       data,
		HeaderFunc: func(m selector.Model, obj any, gdIndex int) string { return "" },
		SelectedFunc: func(m selector.Model, obj any, gdIndex int) string {
			t := obj.(string)
			return common.FontColor(fmt.Sprintf("[%d] %s", gdIndex+1, t), selector.ColorSelected)
		},
		UnSelectedFunc: func(m selector.Model, obj any, gdIndex int) string {
			t := obj.(string)
			return common.FontColor(fmt.Sprintf(" %d. %s", gdIndex+1, t), selector.ColorUnSelected)
		},
		FinishedFunc: func(s any) string {
			return s.(string) + "\n"
		},
		FooterFunc: func(m selector.Model, obj any, gdIndex int) string {
			t := m.Selected().(string)
			return common.FontColor(fmt.Sprint(t), selector.ColorFooter)
		},
	}
}
func fetchChanges() []string {
	status, err := exec.Command("git", "status", "-s").Output()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	arr := strings.Split(string(status), "\n")
	newArr := make([]string, 0)
	for _, v := range arr {
		if len(v) < 3 {
			continue
		}
		newArr = append(newArr, v[3:])
	}
	return newArr
}
func applyGitStash(m *stashUi) {
	if len(m.stagedChanges) < 1 {
		os.Stdout.WriteString("No changes selected")
		os.Exit(0)
	}
	args := make([]string, 0)
	args = append(args, "stash")
	args = append(args, "push")
	args = append(args, "-a")
	args = append(args, m.stagedChanges...)

	err := exec.Command("git", args...).Run()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	os.Stdout.WriteString("changes stashed!")
	os.Exit(0)
}

func initProgram() tea.Model {
	stashed := fetchChanges()
	staged := make([]string, 0)
	sldata := make([]any, 0)
	for _, v := range stashed {
		sldata = append(sldata, v)
	}
	return &stashUi{
		canceled:      false,
		changes:       stashed,
		stagedChanges: staged,
		sl:            makeSelector(sldata),
	}
}

func stash() {
	model := initProgram()
	program := tea.NewProgram(model)
	program.Run()
	if !model.(*stashUi).canceled {
		applyGitStash(model.(*stashUi))
	} else {
		os.Stdout.WriteString("user canceled")
		os.Exit(0)
	}
}
