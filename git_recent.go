package git_recent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
)

type branch struct {
	RelativeTime string
	Name         string
	Author       string
}

func Run(ctx context.Context) {
	// format eg 25 minutes ago|master|Bubunyo Nyavor
	resFormat := `"%(committerdate:relative)|%(refname:short)|%(authorname)"`
	cmd := exec.Command("git", "branch",
		"--sort=-committerdate",
		"--format="+resFormat)

	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	out := strings.Split(string(stdout), "\n")

	branches := []branch{}

	var branchNameMaxLen int

	for _, v := range out {
		v := strings.TrimSpace(v)
		if v == "" {
			continue
		}
		p := strings.Split(v, "|")
		branches = append(branches, branch{p[0], p[1], p[2]})
		if len(p[1]) > branchNameMaxLen {
			branchNameMaxLen = len(p[1])
		}
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0000279C ({{ .RelativeTime | red }}) {{ .Name | green }}",
		Inactive: "  ({{ .RelativeTime | red | faint }}) {{ .Name | cyan | faint }}",
		Selected: "git checkouout {{ .Name | green | cyan }}",
	}

	searcher := func(input string, index int) bool {
		branch := branches[index]
		name := strings.Replace(strings.ToLower(branch.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select git branch:",
		Items:     branches,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	cmd = exec.Command("git", "checkout", branches[i].Name)

	stdout, err = cmd.Output()
	if err != nil {
		fmt.Print(err.Error())
	} else {
		fmt.Print(string(stdout))
	}
}
