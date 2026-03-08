package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 簡單快取
var (
	branchCache   string
	branchExpires time.Time
	cacheMutex    sync.RWMutex
)

// ClearCache 清除快取（用於測試）
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	branchCache = ""
	branchExpires = time.Time{}
}

// GetBranch 獲取 Git 分支（帶快取）
func GetBranch(dir string) string {
	cacheMutex.RLock()
	if time.Now().Before(branchExpires) && branchCache != "" {
		result := branchCache
		cacheMutex.RUnlock()
		return result
	}
	cacheMutex.RUnlock()

	// 檢查是否為 Git 倉庫
	gitPath := ".git"
	if dir != "" {
		gitPath = dir + "/.git"
	}
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		// 嘗試找到 Git 根目錄
		cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
		if err := cmd.Run(); err != nil {
			return ""
		}
	}

	// 獲取當前分支
	cmd := exec.Command("git", "-C", dir, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return ""
	}

	// 檢測是否在 worktree 中
	icon := "⚡"
	worktreeLabel := ""
	gitDirCmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	gitDirOutput, err1 := gitDirCmd.Output()

	gitCommonDirCmd := exec.Command("git", "-C", dir, "rev-parse", "--git-common-dir")
	gitCommonDirOutput, err2 := gitCommonDirCmd.Output()

	if err1 == nil && err2 == nil {
		gitDir := strings.TrimSpace(string(gitDirOutput))
		gitCommonDir := strings.TrimSpace(string(gitCommonDirOutput))

		// Normalize paths to absolute paths for valid comparison
		// git rev-parse can return relative paths (to CWD or to the repo root)
		// We resolve them relative to the repo dir if they are relative
		absGitDir := resolvePath(dir, gitDir)
		absCommonDir := resolvePath(dir, gitCommonDir)

		// 如果 git-dir 和 git-common-dir 不同，表示在 worktree 中
		if absGitDir != absCommonDir {
			icon = "🔀"
			worktreeLabel = " (worktree)"
		}
	}

	result := formatBranch(icon, branch, worktreeLabel)

	// 更新快取
	cacheMutex.Lock()
	branchCache = result
	branchExpires = time.Now().Add(5 * time.Second)
	cacheMutex.Unlock()

	return result
}

// FormatWorktreeBranch 格式化結構化 worktree 資料的分支顯示
func FormatWorktreeBranch(name, branch string) string {
	if branch == "" {
		return ""
	}
	label := " (worktree)"
	if name != "" {
		label = fmt.Sprintf(" (worktree: %s)", name)
	}
	return formatBranch("🔀", branch, label)
}

// formatBranch 統一的分支格式化內部函式
func formatBranch(icon, branch, label string) string {
	return fmt.Sprintf(" %s %s%s", icon, branch, label)
}

// resolvePath resolves a git path to absolute path.
// If path is relative, it is joined with baseDir.
func resolvePath(baseDir, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	// Ensure baseDir is absolute
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		// If we can't get absolute path of base, try our best with what we have
		// or fall back to original logic (string join)
		return filepath.Join(baseDir, path)
	}

	return filepath.Join(absBase, path)
}
