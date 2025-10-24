package git

import (
	"fmt"
	"os"
	"os/exec"
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

		// 如果 git-dir 和 git-common-dir 不同，表示在 worktree 中
		if gitDir != gitCommonDir {
			icon = "🔀"
			worktreeLabel = " (worktree)"
		}
	}

	result := fmt.Sprintf(" %s %s%s", icon, branch, worktreeLabel)

	// 更新快取
	cacheMutex.Lock()
	branchCache = result
	branchExpires = time.Now().Add(5 * time.Second)
	cacheMutex.Unlock()

	return result
}
