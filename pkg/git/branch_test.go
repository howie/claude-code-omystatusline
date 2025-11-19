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

	// 創建 .worktrees/ 目錄
	worktreesDir := filepath.Join(tmpDir, ".worktrees")
	if err := os.Mkdir(worktreesDir, 0755); err != nil {
		t.Fatalf("Failed to create .worktrees dir: %v", err)
	}

	// 創建實際的 worktree
	worktreePath := filepath.Join(worktreesDir, "test-branch")
	runGitCommand(t, tmpDir, "worktree", "add", "-b", "test-branch", worktreePath)

	// 獲取 worktree 的分支資訊
	result := GetBranch(worktreePath)

	// 驗證：實際的 worktree 應該使用 🔀 圖示並有 (worktree) 標籤
	if !strings.Contains(result, "🔀") {
		t.Errorf("Expected 🔀 icon for worktree, got: %s", result)
	}
	if !strings.Contains(result, "(worktree)") {
		t.Errorf("Worktree should have (worktree) label, got: %s", result)
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

// Helper function to run git commands
func runGitCommand(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Git command failed: git %v\nError: %v\nOutput: %s", args, err, output)
	}
}
