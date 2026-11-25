package main

import (
	"testing"

	"maragu.dev/is"
)

func TestHashPort(t *testing.T) {
	// The expected ports come from the port function in versitygw.sh, so the two stay the same.
	tests := []struct {
		dir      string
		name     string
		expected int
	}{
		{dir: "/home/me/app", name: "app", expected: 21334},
		{dir: "/home/me/app", name: "versitygw", expected: 23056},
		{dir: "/home/me/app", name: "versitygw-webui", expected: 22153},
		{dir: "/home/me/app", name: "versitygw-test", expected: 20569},
		{dir: "/home/me/app/.claude/worktrees/feature", name: "app", expected: 22339},
		{dir: "/home/me/app/.claude/worktrees/feature", name: "versitygw", expected: 24367},
		{dir: "/home/me/app/.claude/worktrees/feature", name: "versitygw-webui", expected: 26716},
		{dir: "/home/me/app/.claude/worktrees/feature", name: "versitygw-test", expected: 25323},
	}

	for _, test := range tests {
		t.Run("should return the same port as versitygw.sh for "+test.dir+"/"+test.name, func(t *testing.T) {
			is.Equal(t, test.expected, hashPort(test.dir, test.name))
		})
	}
}
