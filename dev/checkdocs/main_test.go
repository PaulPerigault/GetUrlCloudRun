package main

import "testing"

func TestKnownTask(t *testing.T) {
	tasks := map[string]bool{"verify": true, "infra:test": true, "infra:fmt": true}
	tests := []struct {
		name string
		task string
		want bool
	}{
		{"exact", "verify", true},
		{"exact with colon", "infra:test", true},
		{"wildcard prefix", "infra:", true},
		{"unknown", "deploy", false},
		{"unknown wildcard prefix", "cloud:", false},
		{"partial name is not a prefix", "infra", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := knownTask(tasks, tc.task); got != tc.want {
				t.Errorf("knownTask(%q) = %v, want %v", tc.task, got, tc.want)
			}
		})
	}
}

func TestCheckSnippetTasksAndCommands(t *testing.T) {
	tasks := map[string]bool{"verify": true, "infra:test": true}
	tests := []struct {
		snippet string
		want    int
	}{
		{"task verify", 0},
		{"task infra:*", 0},
		{"task deploy", 1},
		{"geturlcloudrun serve", 0},
		{"geturlcloudrun frobnicate", 1},
	}
	for _, tc := range tests {
		if got := len(checkSnippet("README.md:1", tc.snippet, tasks)); got != tc.want {
			t.Errorf("checkSnippet(%q) = %d problems, want %d", tc.snippet, got, tc.want)
		}
	}
}
