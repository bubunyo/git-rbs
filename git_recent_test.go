package git_recent

import "testing"

func Test_Run(t *testing.T) {
	t.Run("g-recent-run", func(t *testing.T) {
		gt.Run()
	})
}
