package rbs

import (
	"regexp"
	"strings"
	"testing"
)

var branches = []string{"main", "master", "signup", "dawkaka_login", "dawkaka.chat", "webRTC_dawkaka", "pay_two", "pay_one", "socketio_dawkaka", "banda005", "323232"}

func TestSearcher(t *testing.T) {
	var count int
	for key, _ := range branches {
		if searcher("$regex .", key) {
			count++
		}
	}
	if count != len(branches) {
		t.Fatalf("Testing ($regex .): expected %d got %d", len(branches), count)
	}

	//match branch names with only alphabets
	matchedBranches := [3]string{}
	expectedBranches := [3]string{"main", "master", "signup"}
	count = 0
	for key, value := range branches {
		if searcher("$regex ^[a-z]([a-z]+)[a-z]$", key) {
			matchedBranches[count] = value
			count++
		}
	}
	if expectedBranches != matchedBranches {
		t.Fatalf("Testing ($regex ^[a-z]([a-z]+)[a-z]$): expected %v got %v", expectedBranches, matchedBranches)
	}

	//match branch names container numbers
	matchedNumBranches := [2]string{"banda005", "323232"}
	expectedNumBranches := [2]string{"banda005", "323232"}
	count = 0
	for key, value := range branches {
		if searcher("$regex [0-9]", key) {
			matchedNumBranches[count] = value
			count++
		}
	}
	if expectedNumBranches != matchedNumBranches {
		t.Fatalf("Testing ($regex ^[0-9]): expected %v got %v", expectedNumBranches, matchedNumBranches)
	}

	//match branch names with non alphanumberic_ characters
	want := "dawkaka.chat"
	var got string
	for key, value := range branches {
		if searcher("$regex \\W", key) {
			got = value
		}
	}
	if want != got {
		t.Fatalf("Testing ($regex \\W): expected %s got %s", want, got)
	}
}

func searcher(input string, index int) bool {
	branch := branches[index]
	name := strings.ReplaceAll(strings.ToLower(branch), " ", "")
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
