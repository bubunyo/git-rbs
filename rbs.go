package rbs

import (
	"fmt"
	"os/exec"
	"regexp"
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

func Run() {
	// command -  git branch --sort=-committerdate
	// --format="%(committerdate:unix)|%(committerdate:relative)|%(refname:short)|%(authorname)|%(contents:subject)|%(objectname:short)"
	resFormat := `%(objectname:short)|%(refname:short)|%(committerdate:relative)|%(upstream:track)|%(contents:subject)`
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
		branches = append(branches, branch{p[0], p[1], p[2], p[3], p[4]})
		if len(p[1]) > branchNameMaxLen {
			branchNameMaxLen = len(p[1])
		}
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0000279C [{{ .ObjectName | yellow }}] {{ .Name | green }} ({{ .RelativeTime | cyan }}) -{{if ne .Track \"\" }} {{ .Track | magenta }}{{end}} {{ .Subject }}",
		Inactive: "  [{{ .ObjectName | yellow | faint }}] {{ .Name | green | faint  }} ({{ .RelativeTime | cyan | faint }}) -{{if ne .Track \"\" }} {{ .Track | magenta | faint }}{{end}} {{ .Subject | faint }}",
		Selected: "git checkout {{ .Name | green | cyan }}",
	}

	searcher := func(input string, index int) bool {
		branch := branches[index]
		name := strings.ReplaceAll(strings.ToLower(branch.Name), " ", "")
		query := strings.Split(input, " ")

		if query[0] == "$regex" && len(query) > 1 {
			input = query[1]
			reg, err := regexp.Compile(input)
			if err != nil {
				return false
			}
			return reg.MatchString(name)
		}

		input = strings.ReplaceAll(strings.ToLower(input), " ", "")

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
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Print(string(stdout))
	}
}
