package rbs

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
)

type branch struct {
	ObjectName   string
	Name         string
	RelativeTime string
	Track        string
	Subject      string
}

func listBranches() ([]branch, error) {
	resFormat := `%(objectname:short)|%(refname:short)|%(committerdate:relative)|%(upstream:track)|%(contents:subject)`
	cmd := exec.Command("git", "branch",
		"--sort=-committerdate",
		"--format="+resFormat)

	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	out := strings.Split(string(stdout), "\n")
	branches := []branch{}
	for _, v := range out {
		v := strings.TrimSpace(v)
		if v == "" {
			continue
		}
		p := strings.Split(v, "|")
		if len(p) < 5 {
			continue
		}
		branches = append(branches, branch{p[0], p[1], p[2], p[3], p[4]})
	}
	return branches, nil
}

func branchSelectPrompt(label string, branches []branch, selectedTpl string) (int, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0000279C [{{ .ObjectName | yellow }}] {{ .Name | green }} ({{ .RelativeTime | cyan }}) -{{if ne .Track \"\" }} {{ .Track | magenta }}{{end}} {{ .Subject }}",
		Inactive: "  [{{ .ObjectName | yellow | faint }}] {{ .Name | green | faint  }} ({{ .RelativeTime | cyan | faint }}) -{{if ne .Track \"\" }} {{ .Track | magenta | faint }}{{end}} {{ .Subject | faint }}",
		Selected: selectedTpl,
	}

	searcher := func(input string, index int) bool {
		b := branches[index]
		name := strings.Replace(strings.ToLower(b.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     label,
		Items:     branches,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()
	return i, err
}

func Run() {
	branches, err := listBranches()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	i, err := branchSelectPrompt("Select git branch:", branches, "git checkout {{ .Name | green | cyan }}")
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	cmd := exec.Command("git", "checkout", branches[i].Name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s", string(out))
	} else {
		fmt.Print(string(out))
	}
}

// RunWorktree presents an interactive branch selector and automatically adds
// the chosen branch as a new git worktree at ../<branch-name> (a sibling
// directory next to the current repo), replacing forward-slashes in the
// branch name with dashes so that the directory name is always valid.
func RunWorktree() {
	branches, err := listBranches()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	i, err := branchSelectPrompt("Select branch for new worktree:", branches, "git worktree add {{ .Name | green | cyan }}")
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	selected := branches[i]

	// Automatically derive the worktree path: sibling directory named after
	// the branch (slashes replaced with dashes to keep the path component simple).
	dirName := strings.ReplaceAll(selected.Name, "/", "-")
	worktreePath := filepath.Join("..", dirName)

	cmd := exec.Command("git", "worktree", "add", worktreePath, selected.Name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s", string(out))
	} else {
		fmt.Printf("Worktree created at %s\n", worktreePath)
		fmt.Print(string(out))
	}
}
