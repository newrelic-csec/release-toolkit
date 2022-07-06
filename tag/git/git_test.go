package git_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/newrelic/release-toolkit/tag/git"
	"github.com/stretchr/testify/assert"
)

func repoWithTags(t *testing.T, tags ...string) string {
	t.Helper()

	dir := t.TempDir()

	cmds := []string{
		"git init",
		"git config user.email test@user.tld",
		"git config user.name Test",
		"touch a",
		"git add a",
		"git commit -m test",
	}

	for _, t := range tags {
		cmds = append(cmds, fmt.Sprintf("git tag %s", t))
	}

	for _, cmdline := range cmds {
		cmdparts := strings.Fields(cmdline)
		//nolint:gosec // This is a test, we trust hardcoded input.
		cmd := exec.Command(cmdparts[0], cmdparts[1:]...)
		cmd.Dir = dir

		out := strings.Builder{}
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Errorf("%s output:\n%s", cmdline, out.String())
			t.Fatalf("Error bootstraping test git repo: %v", err)
		}
	}

	return dir
}

func TestSource_Tags(t *testing.T) {
	repodir := repoWithTags(t,
		"v1.2.3",
		"v1.3.0",
		"v1.4.0",
		"1.5.0",
		"0.1.1.2",
		"helm-chart-1.3.0",
		"helm-chart-1.3.1",
		"2.0.0-beta",
	)

	for _, tc := range []struct {
		name         string
		opts         []git.OptionFunc
		expectedTags []string
	}{
		{
			name: "Default_Settings",
			opts: nil,
			expectedTags: []string{
				"1.2.3",
				"1.3.0",
				"1.4.0",
				"1.5.0",
				"2.0.0-beta",
			},
		},
		{
			name: "Matching_Leading_v",
			opts: []git.OptionFunc{git.Matching("^v")},
			expectedTags: []string{
				"1.2.3",
				"1.3.0",
				"1.4.0",
			},
		},
		{
			name: "Matching_And_Replacing_Prefix",
			opts: []git.OptionFunc{
				git.Matching("^helm-chart-"),
				git.Replacing("helm-chart-", ""),
			},
			expectedTags: []string{
				"1.3.0",
				"1.3.1",
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src, err := git.NewSource(repodir, tc.opts...)
			if err != nil {
				t.Fatalf("Error creating git source: %v", err)
			}

			tags, err := src.Tags()
			if err != nil {
				t.Fatalf("Error fetching tags: %v", err)
			}

			strTags := make([]string, 0, len(tags))
			for _, t := range tags {
				strTags = append(strTags, t.String())
			}

			assert.ElementsMatchf(t, strTags, tc.expectedTags, "Reported tags do not match")
		})
	}
}