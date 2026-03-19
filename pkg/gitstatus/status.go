package gitstatus

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GitStatusInfo 增強型 Git 狀態資訊
type GitStatusInfo struct {
	IsDirty   bool
	Ahead     int
	Behind    int
	Modified  int
	Added     int
	Deleted   int
	Untracked int
}

// 快取
var (
	statusCache   *GitStatusInfo
	statusExpires time.Time
	statusMutex   sync.RWMutex
)

// ClearCache 清除快取（用於測試）
func ClearCache() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	statusCache = nil
	statusExpires = time.Time{}
}

// Get 取得增強型 Git 狀態（帶 5 秒快取）
func Get(dir string) *GitStatusInfo {
	statusMutex.RLock()
	if time.Now().Before(statusExpires) && statusCache != nil {
		result := statusCache
		statusMutex.RUnlock()
		return result
	}
	statusMutex.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	info := &GitStatusInfo{}

	// 並行取得 porcelain 和 ahead/behind
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		parsePorcelain(ctx, dir, info)
	}()

	go func() {
		defer wg.Done()
		parseAheadBehind(ctx, dir, info)
	}()

	wg.Wait()

	// 更新快取
	statusMutex.Lock()
	statusCache = info
	statusExpires = time.Now().Add(5 * time.Second)
	statusMutex.Unlock()

	return info
}

func parsePorcelain(ctx context.Context, dir string, info *GitStatusInfo) {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		info.IsDirty = true
		xy := line[:2]

		switch {
		case xy[0] == '?' || xy[1] == '?':
			info.Untracked++
		case xy[0] == 'A' || xy[1] == 'A':
			info.Added++
		case xy[0] == 'D' || xy[1] == 'D':
			info.Deleted++
		case xy[0] == 'M' || xy[1] == 'M' || xy[0] == 'R' || xy[1] == 'R':
			info.Modified++
		default:
			info.Modified++ // 其他變更歸為 modified
		}
	}
}

func parseAheadBehind(ctx context.Context, dir string, info *GitStatusInfo) {
	// 取得 ahead
	aheadCmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-list", "--count", "@{upstream}..HEAD")
	if output, err := aheadCmd.Output(); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			info.Ahead = n
		}
	}

	// 取得 behind
	behindCmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-list", "--count", "HEAD..@{upstream}")
	if output, err := behindCmd.Output(); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			info.Behind = n
		}
	}
}

// Format 格式化 Git 狀態為顯示字串
func Format(info *GitStatusInfo) string {
	if info == nil {
		return ""
	}

	var parts []string

	if info.IsDirty {
		parts = append(parts, "*")
	}

	if info.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("↑%d", info.Ahead))
	}
	if info.Behind > 0 {
		parts = append(parts, fmt.Sprintf("↓%d", info.Behind))
	}

	// 檔案統計
	var stats []string
	if info.Modified > 0 {
		stats = append(stats, fmt.Sprintf("!%d", info.Modified))
	}
	if info.Added > 0 {
		stats = append(stats, fmt.Sprintf("+%d", info.Added))
	}
	if info.Deleted > 0 {
		stats = append(stats, fmt.Sprintf("✘%d", info.Deleted))
	}
	if info.Untracked > 0 {
		stats = append(stats, fmt.Sprintf("?%d", info.Untracked))
	}
	if len(stats) > 0 {
		parts = append(parts, strings.Join(stats, ""))
	}

	return strings.Join(parts, "")
}
