package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBranch_MainRepo(t *testing.T) {
	// 清除快取
	ClearCache()

	// 創建臨時測試目錄
	tmpDir := t.TempDir()

	// 初始化 git 倉庫
	runGitCommand(t, tmpDir, "init")
	runGitCommand(t, tmpDir, "config", "user.name", "Test User")
	runGitCommand(t, tmpDir, "config", "user.email", "test@example.com")
	runGitCommand(t, tmpDir, "commit", "--allow-empty", "-m", "Initial commit")

	// 獲取分支資訊
	result := GetBranch(tmpDir)

	// 驗證：主倉庫應該使用 ⚡ 圖示，不應該有 (worktree) 標籤
	if !strings.Contains(result, "⚡") {
		t.Errorf("Expected ⚡ icon for main repo, got: %s", result)
	}
	if strings.Contains(result, "(worktree)") {
		t.Errorf("Main repo should not have (worktree) label, got: %s", result)
	}
	if strings.Contains(result, "🔀") {
		t.Errorf("Main repo should not have 🔀 icon, got: %s", result)
	}
}

func TestGetBranch_MainRepoWithWorktreesDir(t *testing.T) {
	// 清除快取
	ClearCache()

	// 創建臨時測試目錄
	tmpDir := t.TempDir()

	// 初始化 git 倉庫
	runGitCommand(t, tmpDir, "init")
	runGitCommand(t, tmpDir, "config", "user.name", "Test User")
	runGitCommand(t, tmpDir, "config", "user.email", "test@example.com")
	runGitCommand(t, tmpDir, "commit", "--allow-empty", "-m", "Initial commit")

	// 創建 .worktrees/ 目錄（這是 issue #9 的關鍵場景）
	worktreesDir := filepath.Join(tmpDir, ".worktrees")
	if err := os.Mkdir(worktreesDir, 0755); err != nil {
		t.Fatalf("Failed to create .worktrees dir: %v", err)
	}

	// 獲取分支資訊
	result := GetBranch(tmpDir)

	// 驗證：即使有 .worktrees/ 目錄，主倉庫仍應該使用 ⚡ 圖示
	if !strings.Contains(result, "⚡") {
		t.Errorf("Expected ⚡ icon for main repo with .worktrees/ dir, got: %s", result)
	}
	if strings.Contains(result, "(worktree)") {
		t.Errorf("Main repo with .worktrees/ dir should not have (worktree) label, got: %s", result)
	}
	if strings.Contains(result, "🔀") {
		t.Errorf("Main repo with .worktrees/ dir should not have 🔀 icon, got: %s", result)
	}
}

func TestGetBranch_ActualWorktree(t *testing.T) {
	// 清除快取
	ClearCache()

	// 創建臨時測試目錄
	tmpDir := t.TempDir()
	// Resolve symlinks to ensure we have the canonical path (fixes macOS /var vs /private/var issue)
	if resolvedPath, err := filepath.EvalSymlinks(tmpDir); err == nil {
		tmpDir = resolvedPath
	}

	// 初始化 git 倉庫
	runGitCommand(t, tmpDir, "init")
	runGitCommand(t, tmpDir, "config", "user.name", "Test User")
	runGitCommand(t, tmpDir, "config", "user.email", "test@example.com")

	// 創建一個實際的文件以確保倉庫完全初始化
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	runGitCommand(t, tmpDir, "add", "README.md")
	runGitCommand(t, tmpDir, "commit", "-m", "Initial commit")

	// 創建實際的 worktree（放在 repo 外部的獨立 temp 目錄避免 git index 衝突）
	worktreeDir := t.TempDir()
	if resolvedPath, err := filepath.EvalSymlinks(worktreeDir); err == nil {
		worktreeDir = resolvedPath
	}
	worktreePath := filepath.Join(worktreeDir, "test-branch")
	runGitCommand(t, tmpDir, "worktree", "add", "-b", "test-branch", worktreePath)

	// 獲取 worktree 的分支資訊
	result := GetBranch(worktreePath)

	// 驗證：實際的 worktree 應該使用 🔀 圖示並有 (wt) 標籤
	if !strings.Contains(result, "🔀") {
		t.Errorf("Expected 🔀 icon for worktree, got: %s", result)
	}
	if !strings.Contains(result, "(wt)") {
		t.Errorf("Worktree should have (wt) label, got: %s", result)
	}
	if !strings.Contains(result, "test-branch") {
		t.Errorf("Expected branch name 'test-branch' in result, got: %s", result)
	}
	if strings.Contains(result, "⚡") {
		t.Errorf("Worktree should not have ⚡ icon, got: %s", result)
	}
}

func TestGetBranch_NonGitDirectory(t *testing.T) {
	// 清除快取
	ClearCache()

	// 創建臨時測試目錄（不初始化 git）
	tmpDir := t.TempDir()

	// 獲取分支資訊
	result := GetBranch(tmpDir)

	// 驗證：非 git 目錄應該返回空字串
	if result != "" {
		t.Errorf("Expected empty string for non-git directory, got: %s", result)
	}
}

func TestGetBranch_EmptyDirectory(t *testing.T) {
	// 清除快取
	ClearCache()

	// 測試空字串參數
	result := GetBranch("")

	// 應該返回空字串或不崩潰
	if result != "" {
		t.Logf("GetBranch with empty dir returned: %s", result)
	}
}

func TestGetBranch_VerifyGitCommands(t *testing.T) {
	// 清除快取
	ClearCache()

	// 這個測試驗證 git 命令是否在正確的目錄執行
	tmpDir := t.TempDir()

	// 初始化 git 倉庫
	runGitCommand(t, tmpDir, "init")
	runGitCommand(t, tmpDir, "config", "user.name", "Test User")
	runGitCommand(t, tmpDir, "config", "user.email", "test@example.com")
	runGitCommand(t, tmpDir, "commit", "--allow-empty", "-m", "Initial commit")

	// 創建並切換到新分支
	runGitCommand(t, tmpDir, "checkout", "-b", "feature-branch")

	// 獲取分支資訊
	result := GetBranch(tmpDir)

	// 驗證：應該返回正確的分支名稱
	if !strings.Contains(result, "feature-branch") {
		t.Errorf("Expected branch name 'feature-branch' in result, got: %s", result)
	}
}

func TestFormatWorktreeBranch(t *testing.T) {
	tests := []struct {
		name        string
		wtName      string
		branch      string
		wantShort   bool   // true = expect (wt), false = expect (worktree: name)
		wantContain string // branch name should appear in result
	}{
		{
			name:        "name equals branch — smart-omit",
			wtName:      "feature",
			branch:      "feature",
			wantShort:   true,
			wantContain: "feature",
		},
		{
			name:        "branch contains name — smart-omit",
			wtName:      "fix",
			branch:      "hotfix-deploy",
			wantShort:   true,
			wantContain: "hotfix-deploy",
		},
		{
			name:        "name contains branch — smart-omit",
			wtName:      "worktree-fix-story-gen-flow",
			branch:      "fix-story-gen-flow",
			wantShort:   true,
			wantContain: "fix-story-gen-flow",
		},
		{
			name:        "no overlap — full label",
			wtName:      "hotfix",
			branch:      "main",
			wantShort:   false,
			wantContain: "main",
		},
		{
			name:        "empty name — always short",
			wtName:      "",
			branch:      "feature",
			wantShort:   true,
			wantContain: "feature",
		},
		{
			name:        "empty branch — returns empty",
			wtName:      "any",
			branch:      "",
			wantShort:   true, // irrelevant, result is ""
			wantContain: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatWorktreeBranch(tc.wtName, tc.branch)

			if tc.branch == "" {
				if result != "" {
					t.Errorf("empty branch should return empty string, got %q", result)
				}
				return
			}

			if !strings.Contains(result, tc.wantContain) {
				t.Errorf("expected result to contain %q, got %q", tc.wantContain, result)
			}
			if !strings.Contains(result, "🔀") {
				t.Errorf("expected 🔀 icon, got %q", result)
			}
			if tc.wantShort {
				if !strings.Contains(result, "(wt)") {
					t.Errorf("expected short (wt) label, got %q", result)
				}
				if strings.Contains(result, "worktree:") {
					t.Errorf("expected no full worktree label, got %q", result)
				}
			} else {
				if strings.Contains(result, "(wt)") {
					t.Errorf("expected full (worktree: name) label, got %q", result)
				}
				if !strings.Contains(result, "(worktree: "+tc.wtName+")") {
					t.Errorf("expected (worktree: %s) in result, got %q", tc.wtName, result)
				}
			}
		})
	}
}

// Helper function to run git commands
// Clears GIT_* env vars to avoid interference from parent git processes (e.g. pre-commit hooks)
func runGitCommand(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	// Filter out GIT_* env vars that may leak from parent git context
	var cleanEnv []string
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "GIT_") {
			cleanEnv = append(cleanEnv, env)
		}
	}
	cmd.Env = cleanEnv
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Git command failed: git %v\nError: %v\nOutput: %s", args, err, output)
	}
}
