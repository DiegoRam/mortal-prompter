package orchestrator

import (
	"testing"

	"github.com/minimalart/mortal-prompter/internal/git"
	"github.com/minimalart/mortal-prompter/pkg/types"
)

func TestCountFilesInDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected int
	}{
		{
			name:     "empty diff",
			diff:     "",
			expected: 0,
		},
		{
			name: "single file",
			diff: `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
+// new comment`,
			expected: 1,
		},
		{
			name: "multiple files",
			diff: `diff --git a/file1.go b/file1.go
index 1234567..abcdefg 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main
+// comment 1
diff --git a/file2.go b/file2.go
index 1234567..abcdefg 100644
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,4 @@
 package main
+// comment 2
diff --git a/file3.go b/file3.go
index 1234567..abcdefg 100644
--- a/file3.go
+++ b/file3.go
@@ -1,3 +1,4 @@
 package main
+// comment 3`,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countFilesInDiff(tt.diff)
			if result != tt.expected {
				t.Errorf("countFilesInDiff() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestExtractFilesFromDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected []string
	}{
		{
			name:     "empty diff",
			diff:     "",
			expected: []string{},
		},
		{
			name: "single file",
			diff: `diff --git a/internal/config/config.go b/internal/config/config.go
--- a/internal/config/config.go
+++ b/internal/config/config.go
@@ -1,3 +1,4 @@
 package config`,
			expected: []string{"internal/config/config.go"},
		},
		{
			name: "multiple files",
			diff: `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main
diff --git a/pkg/types/types.go b/pkg/types/types.go
--- a/pkg/types/types.go
+++ b/pkg/types/types.go
@@ -1,3 +1,4 @@
 package types`,
			expected: []string{"file1.go", "pkg/types/types.go"},
		},
		{
			name: "new file",
			diff: `diff --git a/newfile.go b/newfile.go
new file mode 100644
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+package main`,
			expected: []string{"newfile.go"},
		},
		{
			name: "deleted file excluded",
			diff: `diff --git a/deleted.go b/deleted.go
deleted file mode 100644
--- a/deleted.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilesFromDiff(tt.diff)
			if len(result) != len(tt.expected) {
				t.Errorf("extractFilesFromDiff() returned %d files, want %d", len(result), len(tt.expected))
				return
			}
			for i, file := range result {
				if file != tt.expected[i] {
					t.Errorf("extractFilesFromDiff()[%d] = %s, want %s", i, file, tt.expected[i])
				}
			}
		})
	}
}

func TestOrchestratorGetState(t *testing.T) {
	orch := &Orchestrator{
		state: types.StateInitializing,
	}

	if orch.GetState() != types.StateInitializing {
		t.Errorf("GetState() = %v, want %v", orch.GetState(), types.StateInitializing)
	}

	orch.state = types.StateRunning
	if orch.GetState() != types.StateRunning {
		t.Errorf("GetState() = %v, want %v", orch.GetState(), types.StateRunning)
	}
}

func TestOrchestratorGetCurrentRound(t *testing.T) {
	orch := &Orchestrator{
		currentRound: 0,
	}

	if orch.GetCurrentRound() != 0 {
		t.Errorf("GetCurrentRound() = %d, want 0", orch.GetCurrentRound())
	}

	orch.currentRound = 5
	if orch.GetCurrentRound() != 5 {
		t.Errorf("GetCurrentRound() = %d, want 5", orch.GetCurrentRound())
	}
}

func TestOrchestratorGetRounds(t *testing.T) {
	rounds := []types.Round{
		{Number: 1, HasIssues: true},
		{Number: 2, HasIssues: false},
	}

	orch := &Orchestrator{
		rounds: rounds,
	}

	result := orch.GetRounds()
	if len(result) != 2 {
		t.Errorf("GetRounds() returned %d rounds, want 2", len(result))
	}

	if result[0].Number != 1 || result[1].Number != 2 {
		t.Errorf("GetRounds() returned unexpected round numbers")
	}
}

func TestBuildResultSuccess(t *testing.T) {
	// Create temp dir for git operations
	tempDir := t.TempDir()

	orch := &Orchestrator{
		rounds: []types.Round{
			{Number: 1, HasIssues: true, GitDiff: "+++ b/file1.go\n"},
			{Number: 2, HasIssues: false, GitDiff: "+++ b/file2.go\n"},
		},
		git: git.New(tempDir),
	}

	result := orch.buildResult(true)

	if !result.Success {
		t.Error("buildResult(true) should set Success to true")
	}

	if result.TotalRounds != 2 {
		t.Errorf("buildResult() TotalRounds = %d, want 2", result.TotalRounds)
	}

	if len(result.FilesModified) != 2 {
		t.Errorf("buildResult() FilesModified count = %d, want 2", len(result.FilesModified))
	}
}

func TestBuildResultFailure(t *testing.T) {
	// Create temp dir for git operations
	tempDir := t.TempDir()

	orch := &Orchestrator{
		rounds: []types.Round{
			{Number: 1, HasIssues: true},
		},
		git: git.New(tempDir),
	}

	result := orch.buildResult(false)

	if result.Success {
		t.Error("buildResult(false) should set Success to false")
	}

	if result.TotalRounds != 1 {
		t.Errorf("buildResult() TotalRounds = %d, want 1", result.TotalRounds)
	}
}
