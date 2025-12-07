package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testRepo represents a temporary git repository for testing.
type testRepo struct {
	dir string
	t   *testing.T
}

// newTestRepo creates a new temporary git repository for testing.
func newTestRepo(t *testing.T) *testRepo {
	t.Helper()

	// Create temp directory
	dir, err := os.MkdirTemp("", "mortal-prompter-git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@mortal-prompter.local")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to configure git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to configure git username: %v", err)
	}

	return &testRepo{dir: dir, t: t}
}

// cleanup removes the temporary repository.
func (r *testRepo) cleanup() {
	os.RemoveAll(r.dir)
}

// createFile creates a file in the test repository.
func (r *testRepo) createFile(name, content string) {
	r.t.Helper()
	path := filepath.Join(r.dir, name)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		r.t.Fatalf("failed to create directories: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		r.t.Fatalf("failed to create file %s: %v", name, err)
	}
}

// run runs a git command in the test repository.
func (r *testRepo) run(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v failed: %v\noutput: %s", args, err, output)
	}
	return string(output)
}

func TestNew(t *testing.T) {
	g := New("/some/path")

	if g.workDir != "/some/path" {
		t.Errorf("expected workDir to be '/some/path', got %q", g.workDir)
	}
}

func TestWorkDir(t *testing.T) {
	g := New("/test/dir")

	if g.WorkDir() != "/test/dir" {
		t.Errorf("expected WorkDir() to return '/test/dir', got %q", g.WorkDir())
	}
}

func TestIsGitRepo_True(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	g := New(repo.dir)

	if !g.IsGitRepo() {
		t.Error("expected IsGitRepo() to return true for a git repository")
	}
}

func TestIsGitRepo_False(t *testing.T) {
	// Create a temp directory that is NOT a git repo
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)

	if g.IsGitRepo() {
		t.Error("expected IsGitRepo() to return false for a non-git directory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit so we have a branch
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	g := New(repo.dir)
	branch, err := g.GetCurrentBranch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default branch could be "main" or "master" depending on git config
	if branch != "main" && branch != "master" {
		t.Errorf("expected branch to be 'main' or 'master', got %q", branch)
	}
}

func TestGetCurrentBranch_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)
	_, err = g.GetCurrentBranch()

	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if err != ErrNotGitRepo {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestHasUncommittedChanges_NoChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	g := New(repo.dir)
	hasChanges, err := g.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hasChanges {
		t.Error("expected HasUncommittedChanges() to return false after clean commit")
	}
}

func TestHasUncommittedChanges_WithUnstagedChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make unstaged change
	repo.createFile("README.md", "# Test Modified")

	g := New(repo.dir)
	hasChanges, err := g.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasChanges {
		t.Error("expected HasUncommittedChanges() to return true with unstaged changes")
	}
}

func TestHasUncommittedChanges_WithStagedChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make staged change
	repo.createFile("README.md", "# Test Staged")
	repo.run("add", "-A")

	g := New(repo.dir)
	hasChanges, err := g.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasChanges {
		t.Error("expected HasUncommittedChanges() to return true with staged changes")
	}
}

func TestHasUncommittedChanges_WithUntrackedFiles(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Add untracked file
	repo.createFile("new-file.txt", "new content")

	g := New(repo.dir)
	hasChanges, err := g.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasChanges {
		t.Error("expected HasUncommittedChanges() to return true with untracked files")
	}
}

func TestGetUnstagedDiff(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Original")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make unstaged change
	repo.createFile("README.md", "# Modified")

	g := New(repo.dir)
	diff, err := g.GetUnstagedDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(diff, "Modified") {
		t.Errorf("expected diff to contain 'Modified', got: %s", diff)
	}
	if !strings.Contains(diff, "Original") {
		t.Errorf("expected diff to contain 'Original', got: %s", diff)
	}
}

func TestGetUnstagedDiff_NoChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	g := New(repo.dir)
	diff, err := g.GetUnstagedDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff != "" {
		t.Errorf("expected empty diff, got: %s", diff)
	}
}

func TestGetStagedDiff(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Original")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make staged change
	repo.createFile("README.md", "# Staged")
	repo.run("add", "-A")

	g := New(repo.dir)
	diff, err := g.GetStagedDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(diff, "Staged") {
		t.Errorf("expected diff to contain 'Staged', got: %s", diff)
	}
}

func TestGetStagedDiff_NoStagedChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make unstaged change (not staged)
	repo.createFile("README.md", "# Modified but not staged")

	g := New(repo.dir)
	diff, err := g.GetStagedDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Staged diff should be empty since changes are not staged
	if diff != "" {
		t.Errorf("expected empty staged diff, got: %s", diff)
	}
}

func TestGetAllDiff(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Original")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make both staged and unstaged changes
	repo.createFile("README.md", "# Staged")
	repo.run("add", "-A")
	repo.createFile("other.txt", "Unstaged")

	g := New(repo.dir)
	diff, err := g.GetAllDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(diff, "Staged") {
		t.Errorf("expected diff to contain 'Staged', got: %s", diff)
	}
}

func TestStageAll(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Create new file and modify existing
	repo.createFile("new-file.txt", "new content")
	repo.createFile("README.md", "# Modified")

	g := New(repo.dir)

	// Before staging, staged diff should be empty
	stagedBefore, _ := g.GetStagedDiff()
	if stagedBefore != "" {
		t.Errorf("expected empty staged diff before StageAll, got: %s", stagedBefore)
	}

	err := g.StageAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After staging, staged diff should have content
	stagedAfter, _ := g.GetStagedDiff()
	if stagedAfter == "" {
		t.Error("expected non-empty staged diff after StageAll")
	}
}

func TestCommit_Success(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Make changes and stage them
	repo.createFile("README.md", "# Modified")
	repo.run("add", "-A")

	g := New(repo.dir)
	err := g.Commit("test commit message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify commit was made
	hasChanges, _ := g.HasUncommittedChanges()
	if hasChanges {
		t.Error("expected no uncommitted changes after commit")
	}
}

func TestCommit_EmptyMessage(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	g := New(repo.dir)
	err := g.Commit("")

	if err == nil {
		t.Error("expected error for empty commit message")
	}
	if err.Error() != "commit message cannot be empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCommit_NoChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit
	repo.createFile("README.md", "# Test")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	g := New(repo.dir)
	err := g.Commit("should fail")

	if err == nil {
		t.Error("expected error when committing with no changes")
	}
	if err != ErrNoChanges {
		t.Errorf("expected ErrNoChanges, got: %v", err)
	}
}

func TestCommit_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create a file so HasUncommittedChanges would work if it were a git repo
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0644)

	g := New(dir)
	err = g.Commit("test message")

	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGetUnstagedDiff_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)
	_, err = g.GetUnstagedDiff()

	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if err != ErrNotGitRepo {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestGetStagedDiff_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)
	_, err = g.GetStagedDiff()

	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if err != ErrNotGitRepo {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestGetAllDiff_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)
	_, err = g.GetAllDiff()

	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if err != ErrNotGitRepo {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestStageAll_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)
	err = g.StageAll()

	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if err != ErrNotGitRepo {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestHasUncommittedChanges_NotGitRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "mortal-prompter-notgit-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	g := New(dir)
	_, err = g.HasUncommittedChanges()

	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if err != ErrNotGitRepo {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestMultipleFilesInDiff(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create initial commit with multiple files
	repo.createFile("file1.go", "package main")
	repo.createFile("file2.go", "package util")
	repo.createFile("sub/file3.go", "package sub")
	repo.run("add", "-A")
	repo.run("commit", "-m", "initial commit")

	// Modify multiple files
	repo.createFile("file1.go", "package main\n// modified")
	repo.createFile("file2.go", "package util\n// modified")
	repo.createFile("sub/file3.go", "package sub\n// modified")

	g := New(repo.dir)
	diff, err := g.GetUnstagedDiff()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that all modified files appear in diff
	if !strings.Contains(diff, "file1.go") {
		t.Error("expected diff to contain file1.go")
	}
	if !strings.Contains(diff, "file2.go") {
		t.Error("expected diff to contain file2.go")
	}
	if !strings.Contains(diff, "file3.go") {
		t.Error("expected diff to contain file3.go")
	}
}

func TestErrorTypes(t *testing.T) {
	// Verify error types are properly defined
	if ErrGitNotInstalled.Error() != "git is not installed or not found in PATH" {
		t.Errorf("unexpected ErrGitNotInstalled message: %v", ErrGitNotInstalled)
	}

	if ErrNotGitRepo.Error() != "not a git repository (or any of the parent directories)" {
		t.Errorf("unexpected ErrNotGitRepo message: %v", ErrNotGitRepo)
	}

	if ErrNoChanges.Error() != "no changes to commit" {
		t.Errorf("unexpected ErrNoChanges message: %v", ErrNoChanges)
	}
}
